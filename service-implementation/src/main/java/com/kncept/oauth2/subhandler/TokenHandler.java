package com.kncept.oauth2.subhandler;

import com.auth0.jwt.JWT;
import com.auth0.jwt.JWTCreator;
import com.auth0.jwt.algorithms.Algorithm;
import com.kncept.oauth2.config.Oauth2StorageConfiguration;
import com.kncept.oauth2.config.authrequest.AuthRequest;
import com.kncept.oauth2.config.parameter.ConfigParameters;
import com.kncept.oauth2.config.session.OauthSession;
import com.kncept.oauth2.crypto.auth.AuthCrypto;
import com.kncept.oauth2.crypto.key.KeyManager;
import com.kncept.oauth2.crypto.key.ManagedKeypair;
import com.kncept.oauth2.operation.response.RenderedContentResponse;
import org.json.simple.JSONObject;

import java.nio.charset.StandardCharsets;
import java.security.interfaces.ECPrivateKey;
import java.security.interfaces.ECPublicKey;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.Map;
import java.util.Optional;

import static com.kncept.oauth2.util.JsonUtils.jsonError;
import static com.kncept.oauth2.util.ParamUtils.required;

public class TokenHandler {

    private final Oauth2StorageConfiguration config;
    private final KeyManager keyManager;

    private final String hostedUrl;

    public TokenHandler(
            Oauth2StorageConfiguration config,
            KeyManager keyManager,
            String hostedUrl
    ) {
        this.config = config;
        this.keyManager = keyManager;
        this.hostedUrl= hostedUrl;
    }

    // https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint
    public RenderedContentResponse token(Map<String, String> params) {
        Optional<String> oauthSessionId = Optional.empty();
        try {
            String grantType = required("grant_type", params);
            // authorization_code

            // requires 'grantType' was authorization_code ?
            String code = required("code", params);
            //code_verifier ? // https://developer.okta.com/docs/reference/api/oidc/#request-parameters-4

            AuthRequest ar = config.authRequestRepository().read(AuthRequest.id(code));
            if (ar == null) {
                return jsonError("No matching auth codes", oauthSessionId);
            }

            // see https://datatracker.ietf.org/doc/html/rfc7636#section-4.6
            if (ar.isPkce()) {
                String codeVerifier = required("code_verifier", params);
                boolean verified = false;
                if("plain".equalsIgnoreCase(ar.getPkceChallengeType())) {
                    verified = ar.getPkceCodeChallenge().equals(codeVerifier);
                } else if ("S256".equalsIgnoreCase(ar.getPkceChallengeType())) {
//                    BASE64URL-ENCODE(SHA256(ASCII(code_verifier))) == code_challenge
                    byte[] b = codeVerifier.getBytes(StandardCharsets.US_ASCII);
                    String expectedCodeChallence = new AuthCrypto().hasher("b64(sha256)").hash(codeVerifier, "");
                    expectedCodeChallence = expectedCodeChallence
                            .replaceAll("=", "") // no padding
                            .replaceAll("\\+", "-") //base64URL replaces + with -
                            .replaceAll("/", "_") //base64URL replaces / with _
                    ;

                    verified = expectedCodeChallence.equals(ar.getPkceCodeChallenge());
                }
                if (!verified) return jsonError("PKCE Verification failure", oauthSessionId);
            }


            OauthSession session = config.oauthSessionRepository().read(ar.getOauthSessionId());
            if (session == null) {
                return jsonError("Session has expired", oauthSessionId);
            }

//            Optional<AuthRequest> authRequest = config.authRequestRepository().lookupByOauthSessionId(authCode.get().oauthSessionId());
//            String responseType = authRequest.map(AuthRequest::responseType).get(); // code vs authorization code
            // redirect_uri for 'authorization code' - same redirect URI as for the original auth request??

            // if the code matches, VEND a JWT token!!
            // https://github.com/auth0/java-jwt

            Instant iat = LocalDateTime.now().toInstant(ZoneOffset.UTC);

            ManagedKeypair keys = keyManager.current();

            Algorithm algorithm = null;
            String alg = keys.keyType();
            if (alg.equals("EC")) {
                algorithm = Algorithm.ECDSA256(
                        (ECPublicKey) keys.keyPair().getPublic(),
                        (ECPrivateKey) keys.keyPair().getPrivate());
            } else if (alg.equals("RSA")) {
                algorithm = Algorithm.RSA256(
                        (RSAPublicKey) keys.keyPair().getPublic(),
                        (RSAPrivateKey) keys.keyPair().getPrivate());
            } else throw new UnsupportedOperationException("Unable to handle a key algorithm of " + alg);
            JWTCreator.Builder jwtBuilder = JWT.create()
                    .withIssuer(hostedUrl)
                    .withSubject(session.getRef().toString())
                    .withIssuedAt(iat)
                    .withExpiresAt(iat.plusSeconds(sessionDurationInSeconds()))
                    ;
            if (ar.getNonce() != null && ar.getNonce().isPresent()) {
                jwtBuilder.withClaim("nonce", ar.getNonce().get());
            }

            String token = jwtBuilder
                    .sign(algorithm);

            JSONObject jwt = new JSONObject();
            jwt.put("token_type", "Bearer");
            jwt.put("id_token", token);
            jwt.put("expires_in", sessionDurationInSeconds());
            //        jwt.put("refresh_token", "xxxx")

            // needs to be json
            return new RenderedContentResponse(200, jwt.toJSONString(), "application/json", oauthSessionId, false)
                    .addHeader("Cache-Control", "no-store")
                    .addHeader("Pragma", "no-cache");

        } catch (RuntimeException e) {
            return jsonError(e.getMessage(), oauthSessionId);
        }
    }



    int sessionDurationInSeconds() {
        return Integer.parseInt(ConfigParameters.sessionDuration.get(config.parameterRepository()));
    }

}
