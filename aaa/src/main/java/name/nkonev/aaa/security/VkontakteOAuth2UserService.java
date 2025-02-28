package name.nkonev.aaa.security;

import name.nkonev.aaa.config.properties.AaaProperties;
import name.nkonev.aaa.config.properties.ConflictResolveStrategy;
import name.nkonev.aaa.dto.UserAccountDetailsDTO;
import name.nkonev.aaa.entity.jdbc.UserAccount;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.oauth2.client.userinfo.DefaultOAuth2UserService;
import org.springframework.security.oauth2.client.userinfo.OAuth2UserRequest;
import org.springframework.security.oauth2.client.userinfo.OAuth2UserService;
import org.springframework.security.oauth2.core.OAuth2AuthenticationException;
import org.springframework.security.oauth2.core.user.OAuth2User;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;

import java.util.*;

import static name.nkonev.aaa.Constants.VKONTAKTE_LOGIN_PREFIX;

@Service
public class VkontakteOAuth2UserService extends AbstractOAuth2UserService implements OAuth2UserService<OAuth2UserRequest, OAuth2User> {

    @Autowired
    private DefaultOAuth2UserService delegate;

    @Autowired
    private AaaProperties aaaProperties;

    private static final Logger LOGGER = LoggerFactory.getLogger(VkontakteOAuth2UserService.class);

    @Override
    public OAuth2User loadUser(OAuth2UserRequest userRequest) throws OAuth2AuthenticationException {
        OAuth2User oAuth2User = delegate.loadUser(userRequest);

        var map = oAuth2User.getAttributes();

        UserAccountDetailsDTO processUserResponse = process(map, userRequest);

        return processUserResponse;
    }

    private Map<String, Object> getMeaningfulMap(Map<String, Object> map) {
        List l = (List) map.get("response");
        Map<String, Object> m = (Map<String, Object>) l.get(0);
        return m;
    }

    @Override
    protected String getId(Map<String, Object> map) {
        Map<String, Object> m = getMeaningfulMap(map);

        return ((Integer) m.get("id")).toString();
    }

    @Override
    protected String getLogin(Map<String, Object> map) {
        Map<String, Object> m = getMeaningfulMap(map);

        String firstName = (String) m.get("first_name");
        String lastName = (String) m.get("last_name");
        String login = "";
        if (firstName!=null) {
            firstName = firstName.trim();
            login += firstName;
            login += " ";
        }
        if (lastName!=null) {
            lastName = lastName.trim();
            login += lastName;
        }
        Assert.hasLength(login, "vkontakte name cannot be null");
        login = login.trim();
        return login;
    }

    @Override
    protected Logger logger() {
        return LOGGER;
    }

    @Override
    protected String getOAuth2Name() {
        return OAuth2Providers.VKONTAKTE;
    }

    @Override
    protected Optional<UserAccount> findByOAuth2Id(String oauthId) {
        return userAccountRepository.findByVkontakteId(oauthId);
    }

    @Override
    protected UserAccountDetailsDTO setOAuth2IdToPrincipal(UserAccountDetailsDTO principal, String oauthId) {
        return principal.withOauth2Identifiers(principal.getOauth2Identifiers().withVkontakteId(oauthId));
    }

    @Override
    protected UserAccount setOAuth2IdToEntity(Long id, String oauthId) {
        UserAccount userAccount = userAccountRepository.findById(id).orElseThrow();
        userAccount = userAccount.withOauthIdentifiers(userAccount.oauth2Identifiers().withVkontakteId(oauthId));
        userAccount = userAccountRepository.save(userAccount);
        return userAccount;
    }

    @Override
    protected UserAccount buildEntity(String oauthId, String login, Map<String, Object> oauthResourceServerResponse, Set<String> roles) {
        UserAccount userAccount = userAccountConverter.buildUserAccountEntityForVkontakteInsert(oauthId, login);
        LOGGER.info("Built {} user id={} login='{}'", getOAuth2Name(), oauthId, login);

        return userAccount;

    }

    @Override
    protected String getConflictPrefix() {
        return VKONTAKTE_LOGIN_PREFIX;
    }

    @Override
    protected ConflictResolveStrategy getConflictResolveStrategy() {
        return aaaProperties.vkontakte().resolveConflictsStrategy();
    }
}
