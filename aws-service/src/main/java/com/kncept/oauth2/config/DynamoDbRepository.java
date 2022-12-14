package com.kncept.oauth2.config;

import com.kncept.oauth2.config.annotation.OidcExpiryTime;
import com.kncept.oauth2.config.annotation.OidcId;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.*;

import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Collectors;

public class DynamoDbRepository<T> {
    private final Class<T> valueInterface;
    public final DynamoDbClient client;
    public final String tableName;

    private List<Method> methods;
    private String idFieldName;
    private Optional<String> ttlField;

    public DynamoDbRepository(Class<T> valueInterface, String tableName) {
        this(valueInterface, DynamoDbClient.create(), tableName);
    }

    public DynamoDbRepository(
            Class<T> valueInterface,
            DynamoDbClient client,
            String tableName
    ) {
        this.valueInterface = valueInterface;
        this.client = client;
        this.tableName = tableName;

        if (!valueInterface.isInterface()) throw new RuntimeException("Must be an interface");
    }

    public long epochSecondsExpiry(long secondsDuration) {
        long epochSecond = System.currentTimeMillis() / 1000L;
        return epochSecond + secondsDuration;
    }

    // add methods from interface (and superinterfaces)
    // and filter for non-void returns with 0 input paramaters
    private List<Method> methods() {
    	if (methods == null) {
    		List<Method> allMethods = allMethods(valueInterface);
    		methods = new ArrayList<>();
    		// now filter them;
    		for(Method m: allMethods) {
    			if (void.class != m.getReturnType() && m.getParameterCount() == 0)
    				methods.add(m);
    		}

    	}
    	return methods;
    }
    private List<Method> allMethods(Class<?> c) {
    	List<Method> methods = new ArrayList<>();
    	methods.addAll(Arrays.asList(c.getDeclaredMethods()));
    	for(Class<?> intf: c.getInterfaces()) {
    		methods.addAll(allMethods(intf));
    	}
    	return methods;
    }
    private String idFieldName() {
        if (idFieldName == null) {
            for (Method m : methods()) {
                if (m.getAnnotation(OidcId.class) != null) {
                	if (idFieldName != null)
                	    throw new IllegalStateException("Multiple ID fields");
                	if (!String.class.isAssignableFrom(m.getReturnType()))
                	    throw new IllegalStateException("Key must be string");
                    idFieldName = m.getName();
                }
            }
            if (idFieldName == null) throw new IllegalStateException("Must have an OIDC Id field: " + valueInterface.getSimpleName());
        }
        return idFieldName;
    }
    private Optional<String> ttlField() {
        if (ttlField == null) {
            for (Method m : methods()) {
                if (m.getAnnotation(OidcExpiryTime.class) != null) {
                    if (ttlField != null)
                        throw new IllegalStateException("Multiple Expiry fields");
                    if (!(long.class.isAssignableFrom(m.getReturnType()) || Long.class.isAssignableFrom(m.getReturnType())))
                        throw new IllegalStateException("Expiry must be long (UTC timestamp)");
                    ttlField = Optional.of(m.getName());
                }
            }
            if (ttlField == null) ttlField = Optional.empty();
        }
        return ttlField;
    }

    // https://github.com/awsdocs/aws-doc-sdk-examples/blob/main/javav2/example_code/dynamodb/src/main/java/com/example/dynamodb/CreateTable.java
    public synchronized void createTableIfNotExists() {
        TableStatus tableStatus = null;
        DescribeTableRequest describeTable = DescribeTableRequest.builder()
                .tableName(tableName)
                .build();
        try {
            DescribeTableResponse describeTableResponse = client.describeTable(describeTable);
            tableStatus = describeTableResponse.table().tableStatus();
        } catch (ResourceNotFoundException rnf) { }
        if (tableStatus == null || !(tableStatus.equals(TableStatus.ACTIVE) || tableStatus.equals(TableStatus.CREATING))) {
            System.out.println(getClass().getSimpleName() + " creating table " + tableName);
            client.createTable(CreateTableRequest.builder()
                    .tableName(tableName)
                    .keySchema(KeySchemaElement.builder()
                            .attributeName(idFieldName())
                            .keyType(KeyType.HASH)
                            .build())
                    .attributeDefinitions(
                        new AttributeDefinition[] {AttributeDefinition.builder()
                                .attributeName(idFieldName())
                                .attributeType(ScalarAttributeType.S)
                                .build()}
                    )
                    .billingMode(BillingMode.PAY_PER_REQUEST)
                    .build());
        }
        while (tableStatus == null || !tableStatus.equals(TableStatus.ACTIVE)) {
            try { Thread.sleep(5); }
            catch (InterruptedException e) { }
            try {
                DescribeTableResponse describeTableResponse = client.describeTable(describeTable);
                tableStatus = describeTableResponse.table().tableStatus();
            } catch (ResourceNotFoundException rnf) { }
        }

        ttlField().ifPresent(ttlFieldName -> {
            DescribeTimeToLiveResponse ttlResponse = client.describeTimeToLive(DescribeTimeToLiveRequest.builder()
                    .tableName(tableName)
                    .build());
            TimeToLiveStatus ttlStatus = ttlResponse.timeToLiveDescription().timeToLiveStatus();
            if (ttlStatus == null || !(ttlStatus.equals(TimeToLiveStatus.ENABLED) || ttlStatus.equals(TimeToLiveStatus.ENABLING))) {
                System.out.println(getClass().getSimpleName() + " updating ttl " + tableName);
                client.updateTimeToLive(UpdateTimeToLiveRequest.builder()
                        .tableName(tableName)
                        .timeToLiveSpecification(TimeToLiveSpecification.builder()
                                .attributeName(ttlFieldName)
                                .enabled(true)
                                .build())
                        .build());
            }
        });

    }


    // read data from ddb
    T reflectiveItemConverter(Map<String, AttributeValue> item) {
        if (item == null || item.isEmpty()) return null;
        ClassLoader classLoader = Thread.currentThread().getContextClassLoader();
        if (classLoader == null) classLoader = getClass().getClassLoader();
        return (T) Proxy.newProxyInstance(classLoader, new Class[] {valueInterface}, new InvocationHandler() {
            @Override
            public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
                String methodName = method.getName();
                AttributeValue av = item.get(methodName);
                return fromAttributeValue(av, method);
            }
        });
    }
    public Object fromAttributeValue(AttributeValue av, Method typeDetails) throws ClassNotFoundException {
    	Class<?> type = typeDetails.getReturnType();
    	// unroll optional
    	if (Optional.class.isAssignableFrom(type)) {
    		try {
    			if (av.nul()) return Optional.empty();
    		} catch (NullPointerException e) {
    			// means it's not null... ugh
    		}

    		ParameterizedType genericReturnType = (ParameterizedType)typeDetails.getGenericReturnType();
    		Type[] typeArgs = genericReturnType.getActualTypeArguments();
    		String optionalGenericTypeParamName = typeArgs[0].getTypeName(); // eg: "java.lang.String" ... ugh
    		return Optional.of(fromAttributeValue(av, Class.forName(optionalGenericTypeParamName)));

    	}
    	return fromAttributeValue(av, type);
    }
    public Object fromAttributeValue(AttributeValue av, Class<?> type) {
    	try {
			if (av.nul()) return null;
		} catch (NullPointerException e) {
			// means it's not null... ugh
		}
    	if (String.class.isAssignableFrom(type)) return av.s();
    	if (Boolean.class.isAssignableFrom(type)) return av.bool();
    	if (boolean.class.isAssignableFrom(type)) return av.bool();
        if (Long.class.isAssignableFrom(type)) return Long.valueOf(av.n());
        if (boolean.class.isAssignableFrom(type)) return Long.valueOf(av.n());

    	throw new RuntimeException("Unable to deconvert field of type " + type.getName());
    }

    // write data to ddb
    // n.b. primitive types will be autoboxed
    public AttributeValue toAttributeValue(Object value) {
    	if (value == null) return AttributeValue.builder().nul(true).build();
    	if (value instanceof Optional) {
    		if (((Optional) value).isEmpty()) return AttributeValue.builder().nul(true).build();
    		value = ((Optional) value).get();
    	}
    	if (value instanceof String) return AttributeValue.builder().s((String)value).build();
    	if (value instanceof Boolean) return AttributeValue.builder().bool((Boolean)value).build();
    	if (value instanceof Long) return AttributeValue.fromN(value.toString());
    	throw new RuntimeException("Unable to convert value of type" + value.getClass().getSimpleName());
    }
    public Map<String, AttributeValue> convert(T value) {
    	try {
	        Map<String, AttributeValue> ddbValue = new TreeMap<>();
	        for(Method m: methods()) {
	        	Object fieldValue = m.invoke(value);
	        	ddbValue.put(m.getName(), toAttributeValue(fieldValue));
	        }
	        return ddbValue;
    	} catch (IllegalAccessException e) {
    		throw new RuntimeException(e);
		} catch (IllegalArgumentException e) {
			 throw new RuntimeException(e);
		} catch (InvocationTargetException e) {
			throw new RuntimeException(e);
		}
    }

    public void write(T value) {
        client.putItem(PutItemRequest.builder()
                .tableName(tableName)
                .item(convert(value))
                .build());
    }
    public void delete(String key)  {
        client.deleteItem(DeleteItemRequest.builder()
                .key(Map.of(idFieldName(), AttributeValue.fromS(key)))
                .build());
    }
    public T findById(String key) {
        try {
            GetItemResponse response = client.getItem(GetItemRequest.builder()
                    .tableName(tableName)
                    .key(Map.of(idFieldName(), AttributeValue.fromS(key)))
                    .build());
            return reflectiveItemConverter(response.item());
        } catch (ResourceNotFoundException rnf) {
            return null;
        }
    }

    public List<T> list() {
        ScanResponse scanResponse = client.scan(ScanRequest.builder().tableName(tableName).build());
        return scanResponse.items().stream().map(this::reflectiveItemConverter).collect(Collectors.toList());
    }

}
