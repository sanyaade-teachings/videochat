package name.nkonev.aaa.utils;

import java.util.function.Supplier;

public abstract class NullUtils {
    public static <T> T getOrNull(Supplier<T> supplier) {
        try {
            return supplier.get();
        } catch (NullPointerException e) {
            return null;
        }
    }

    public static <T> T getOrNullWrapException(CheckedExceptionSupplier<T> supplier) {
        try {
            return supplier.get();
        } catch (NullPointerException e) {
            return null;
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

}