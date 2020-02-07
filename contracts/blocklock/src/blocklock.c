#include <runtime.h>

// UNLOCK_HEIGHT must be defined at compile time.

void unlock() {
        if (ext_get_height() < UNLOCK_HEIGHT) {
                ext_revert("locked", 7);
        }
        ext_self_destruct();
}

int unlock_height() {
        return UNLOCK_HEIGHT;
}
