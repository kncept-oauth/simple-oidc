package com.kncept.oauth2.config;

import com.kncept.oauth2.config.authrequest.AuthRequest;
import com.kncept.oauth2.config.authrequest.SimpleAuthRequest;
import com.kncept.oauth2.config.client.Client;
import com.kncept.oauth2.config.client.SimpleClient;
import com.kncept.oauth2.config.user.SaltedUser;
import com.kncept.oauth2.config.user.SimpleSaltedUser;
import com.kncept.oauth2.config.user.SimpleUser;
import com.kncept.oauth2.config.user.User;

import org.junit.jupiter.api.Test;
import software.amazon.awssdk.services.dynamodb.model.AttributeValue;

import java.util.Map;
import java.util.Optional;
import java.util.UUID;

import static org.junit.jupiter.api.Assertions.*;

public class DynamoDbRepositoryTest {

    @Test
    public void canConvertClient() {
        DynamoDbRepository repository = new DynamoDbRepository<Client>(Client.class, null, "KnceptOidcClientRepository");
        String clientId = UUID.randomUUID().toString();

        Client original = new SimpleClient(clientId);
        Map<String, AttributeValue> converted = repository.convert(original);
        assertNotNull(converted);
        
        assertTrue(converted.containsKey("clientId"));
        assertEquals(clientId, converted.get("clientId").s());
        assertTrue(converted.containsKey("enabled"));
        assertEquals(original.enabled(), converted.get("enabled").bool());

        assertFalse(converted.containsKey("id"));
        
        Client reconstitued = (Client)repository.reflectiveItemConverter(converted);
        assertEquals(original.clientId(), reconstitued.clientId());
        assertEquals(original.enabled(), reconstitued.enabled());
    }
    
    @Test
    public void canConvertAuthRequest() {
    	DynamoDbRepository repository = new DynamoDbRepository(AuthRequest.class, null, "KnceptOidcAuthRequestRepository");
        String oauthSessionId = UUID.randomUUID().toString();
        AuthRequest original = new SimpleAuthRequest(
        		"oauthSessionId string",
                Optional.of("state string"),
                Optional.empty(), // nonce,
                "redirectUri string",
                "clientId string",
                "responseType string",
                0
        		);
        Map<String, AttributeValue> converted = repository.convert(original);
        AuthRequest reconstitued = (AuthRequest)repository.reflectiveItemConverter(converted);
        assertEquals(original.oauthSessionId(), reconstitued.oauthSessionId());
        assertEquals(original.state(), reconstitued.state());
        assertEquals(original.nonce(), reconstitued.nonce());
        assertEquals(original.redirectUri(), reconstitued.redirectUri());
        assertEquals(original.clientId(), reconstitued.clientId());
        assertEquals(original.responseType(), reconstitued.responseType());
    }
    
    @Test
    public void typeConversions() {
    	DynamoDbRepository repository = new DynamoDbRepository(User.class, null, null);
    	AttributeValue av = null;
    	
    	av = repository.toAttributeValue("stringValue");
    	assertEquals("stringValue", av.s());
    	
    	av = repository.toAttributeValue(true);
    	assertEquals(true, av.bool());
    	
    	av = repository.toAttributeValue(null);
    	assertEquals(true, av.nul());
    	
    	av = repository.toAttributeValue(Optional.empty());
    	assertEquals(true, av.nul());
    	
    	av = repository.toAttributeValue(Optional.of("stringOption"));
    	assertEquals("stringOption", av.s());
    }
    
    @Test
    public void typeDeconversions() throws Exception {
    	DynamoDbRepository repository = new DynamoDbRepository(User.class, null, null);
    	AttributeValue av = null;
    	Object value = null;
    	
    	av = AttributeValue.fromS("stringValue");
    	value = repository.fromAttributeValue(av, AuthRequest.class.getDeclaredMethod("oauthSessionId")); // method returning String
    	assertEquals("stringValue", value);
    	
    	av = AttributeValue.fromBool(true);
    	value = repository.fromAttributeValue(av, Client.class.getDeclaredMethod("enabled")); // method returning String
    	assertEquals(true, value);

    	av = AttributeValue.fromNul(true);
    	value = repository.fromAttributeValue(av, AuthRequest.class.getDeclaredMethod("oauthSessionId")); // method returning String
    	assertEquals(null, value);
    	
    	av = AttributeValue.fromNul(true);
    	value = repository.fromAttributeValue(av, AuthRequest.class.getDeclaredMethod("state")); // method returning Optional<String>
    	assertEquals(Optional.empty(), value);
    	
    	av = AttributeValue.fromS("stringOption");
    	value = repository.fromAttributeValue(av, AuthRequest.class.getDeclaredMethod("state")); // method returning Optional<String>
    	assertEquals(Optional.of("stringOption"), value);
    }

    @Test
    public void canConvertSimpleUser() {
        DynamoDbRepository repository = new DynamoDbRepository<User>(User.class, null, "KnceptOidcSimpleUserRepository");
        String randomUserId = UUID.randomUUID().toString();
        SimpleUser original = new SimpleUser(randomUserId, "simpleusername");
        Map<String, AttributeValue> converted = repository.convert(original);
        User reconstitued = (User)repository.reflectiveItemConverter(converted);
        assertEquals(randomUserId, reconstitued.userId());
    }

    @Test
    public void canConvertSimpleSaltedUser() {
        DynamoDbRepository repository = new DynamoDbRepository<SaltedUser>(SaltedUser.class, null, "KnceptOidcSimpleSaltedUserRepository");
        String randomUserId = UUID.randomUUID().toString();
        SimpleSaltedUser original = new SimpleSaltedUser(
                randomUserId,
                "simpleusername",
                "salt",
                "precomputedHash",
                null
        );
        Map<String, AttributeValue> converted = repository.convert(original);
        SaltedUser reconstitued = (SaltedUser)repository.reflectiveItemConverter(converted);
        assertEquals(randomUserId, reconstitued.userId());
        assertNull(reconstitued.hashAlgorithm());
        assertEquals("salt", reconstitued.salt());
    }
}
