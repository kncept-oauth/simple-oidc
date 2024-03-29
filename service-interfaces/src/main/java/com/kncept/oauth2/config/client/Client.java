package com.kncept.oauth2.config.client;


import com.kncept.oauth2.entity.EntityId;
import com.kncept.oauth2.entity.IdentifiedEntity;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Data
public class Client implements IdentifiedEntity {
    public static String EntityType = "client";

    public static EntityId id(String value) {
        return EntityId.parse(EntityType, value);
    }

    private EntityId id; // oauth client id (prefixed 'client/' as per entity id)
    private String secret; // client secret for OAuth2
    boolean enabled;
    boolean requirePkce; // if PKCE is _required_ for the 'authorization' flow
    String[] endpoints; // allowed redirect_uri destinations for this endpoint

    @Override
    public EntityId getRef() {
        return id;
    }

    @Override
    public LocalDateTime getExpiry() {
        return null;
    }

    @Override
    public IdentifiedEntity clone() {
        try {
            return (IdentifiedEntity) super.clone();
        } catch (CloneNotSupportedException e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public LocalDateTime getWhen() {
        return null;
    }

    @Override
    public void validate() {
        if (secret == null) throw new IllegalStateException();
        if (endpoints == null) throw new IllegalStateException();
    }
}
