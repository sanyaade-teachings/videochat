package name.nkonev.aaa.tasks;

import name.nkonev.aaa.repository.jdbc.UserAccountRepository;
import name.nkonev.aaa.security.AaaUserDetailsService;
import name.nkonev.aaa.services.EventService;
import net.javacrumbs.shedlock.spring.annotation.SchedulerLock;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.data.domain.PageRequest;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

@ConditionalOnProperty("custom.schedulers.user-online.enabled")
@Service
public class UserOnlineTask {

    @Autowired
    private UserAccountRepository userAccountRepository;

    @Autowired
    private AaaUserDetailsService aaaUserDetailsService;

    @Autowired
    private EventService eventService;

    @Value("${custom.schedulers.user-online.batch-size}")
    private int userOnlineBatchSize;

    private static final Logger LOGGER = LoggerFactory.getLogger(UserOnlineTask.class);

    @Scheduled(cron = "${custom.schedulers.user-online.cron}")
    @SchedulerLock(name = "userOnlineTask")
    public void scheduledTask() {
        final int pageSize = userOnlineBatchSize;
        LOGGER.debug("User online task start, userOnlineBatchSize={}", userOnlineBatchSize);
        var count = userAccountRepository.count();
        var pages = (count / pageSize) + ((count > pageSize && count % pageSize > 0) ? 1 : 0) + 1;

        for (int i = 0; i < pages; i++) {
            var chunk = userAccountRepository.findAll(PageRequest.of(i, pageSize));
            var usersOnline = aaaUserDetailsService.getUsersOnlineByUsers(chunk.getContent());
            eventService.notifyOnlineChanged(usersOnline);
        }
        LOGGER.debug("User online task finish");
    }
}