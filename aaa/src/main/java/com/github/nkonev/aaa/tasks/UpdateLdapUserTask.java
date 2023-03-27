package com.github.nkonev.aaa.tasks;

import com.github.nkonev.aaa.repository.jdbc.UserAccountRepository;
import net.javacrumbs.shedlock.spring.annotation.SchedulerLock;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.data.domain.PageRequest;
import org.springframework.ldap.core.LdapOperations;
import org.springframework.ldap.query.LdapQueryBuilder;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

@ConditionalOnProperty("custom.ldap.auth.enabled")
@Service
public class UpdateLdapUserTask {

    @Autowired
    private UserAccountRepository userAccountRepository;

    @Value("${custom.ldap.user-update-batch-size}")
    private int batchSize;

    @Autowired
    private LdapOperations ldapOperations;

    @Value("${custom.ldap.auth.base:}")
    private String base;

    @Value("${custom.ldap.auth.filter:}")
    private String filter;

    private static final Logger LOGGER = LoggerFactory.getLogger(UpdateLdapUserTask.class);

    @Scheduled(cron = "${custom.ldap.user-update-cron}") // TODO use unified tasks level in application.yml
    @SchedulerLock(name = "updateLdapUserTask")
    public void scheduledTask() {
        final int pageSize = batchSize;
        LOGGER.debug("Update LDAP user task start, batchSize={}", batchSize);
        var count = userAccountRepository.count();
        var pages = (count / pageSize) + ((count > pageSize && count % pageSize > 0) ? 1 : 0) + 1;

        // TODO iterate over ldap repository
        for (int i = 0; i < pages; i++) {
            var chunk = userAccountRepository.findAll(PageRequest.of(i, pageSize)); // TODO filter by creation_type == LDAP
            var lq = LdapQueryBuilder.query().base(base).filter(filter);
            ldapOperations.find(lq)
        }
        LOGGER.debug("User online task finish");
    }
}
