#ifndef RUNTIME_TEST_H
#define RUNTIME_TEST_H

#define SUCCESS 0

#define INC(val) ((val) + 1)

#define GET_HEIGHT_EXP 1001
#define GET_HEIGHT_ERR INC(SUCCESS)

#define GET_SENDER_ERR INC(GET_HEIGHT_ERR)

#define GET_AMOUNT_EXP 5001
#define GET_AMOUNT_ERR INC(GET_SENDER_ERR)

#define GET_ENTRY_HASH_ERR INC(GET_AMOUNT_ERR)

#define GET_TIMESTAMP_EXP 1575938086
#define GET_TIMESTAMP_ERR INC(GET_ENTRY_HASH_ERR)

#endif // RUNTIME_TEST_H
