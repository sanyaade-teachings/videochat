package name.nkonev.aaa.utils;

import java.time.LocalDateTime;
import java.time.ZoneId;
import java.time.ZoneOffset;
import java.util.Date;

/**
 * Created by nkonev on 15.05.17.
 */
public class TimeUtil {
    public static LocalDateTime getNowUTC() {
        return LocalDateTime.now(ZoneOffset.UTC);
    }

    public static LocalDateTime convertToLocalDateTime(Date dateToConvert) {
        return dateToConvert.toInstant()
                .atZone(ZoneId.from(ZoneOffset.UTC))
                .toLocalDateTime();
    }
}
