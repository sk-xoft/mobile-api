# SQL catalog: auth

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## auth.ConsumeActionToken

Source: [internal/platform/mysql/auth/secure_repository.go:31](../../../internal/platform/mysql/auth/secure_repository.go#L31)

```sql
SELECT user_id FROM mobile_auth_tokens WHERE token_hash=? AND purpose=? AND expires_at>NOW()
```

```sql
SELECT MOBILE_UUID,USERNAME,EMAIL,USER_STATUS FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE
```

```sql
SELECT token_hash FROM mobile_auth_tokens WHERE token_hash=? AND purpose=? AND expires_at>NOW() FOR UPDATE
```

```sql
UPDATE m_mobile_users SET USER_STATUS='A',UPDATE_DATE=NOW(),UPDATE_BY=? WHERE MOBILE_UUID=?
```

```sql
UPDATE m_mobile_users SET PASSWORD=? WHERE MOBILE_UUID=?
```

```sql
UPDATE t_session SET IS_ONLINE='NO' WHERE USER_ID=?
```

```sql
DELETE FROM mobile_auth_tokens WHERE user_id=? AND purpose=?
```

## auth.CreateActionToken

Source: [internal/platform/mysql/auth/secure_repository.go:11](../../../internal/platform/mysql/auth/secure_repository.go#L11)

```sql
SELECT MOBILE_UUID FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE
```

```sql
DELETE FROM mobile_auth_tokens WHERE user_id=? AND purpose=?
```

```sql
INSERT INTO mobile_auth_tokens(token_hash,user_id,purpose,expires_at) VALUES(?,?,?,?)
```

## auth.CreateSecureSession

Source: [internal/platform/mysql/auth/secure_repository.go:84](../../../internal/platform/mysql/auth/secure_repository.go#L84)

```sql
SELECT USER_STATUS FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE
```

```sql
UPDATE t_session SET IS_ONLINE='NO' WHERE USER_ID=? AND IS_ONLINE='YES'
```

```sql
INSERT INTO t_session (CLIENT_SESSION_ID,USER_ID,LOGIN_TIME,DEVICE_MODEL,TOKEN,CREATE_DATE,CREATE_BY,UPDATE_DATE,UPDATE_BY,CORRELATIONID,IS_ONLINE) VALUES(?,?,?,'MOBILE',?,?,?,?,?,?,'YES')
```

```sql
INSERT INTO mobile_auth_sessions(session_id,refresh_hash,expires_at) VALUES(?,?,?)
```

## auth.FindRefreshSession

Source: [internal/platform/mysql/auth/secure_repository.go:110](../../../internal/platform/mysql/auth/secure_repository.go#L110)

```sql
SELECT s.CLIENT_SESSION_ID,s.USER_ID FROM t_session s INNER JOIN mobile_auth_sessions a ON a.session_id=s.CLIENT_SESSION_ID INNER JOIN m_mobile_users u ON u.MOBILE_UUID=s.USER_ID WHERE a.refresh_hash=? AND a.expires_at>NOW() AND s.IS_ONLINE='YES' AND u.USER_STATUS='A' LIMIT 1
```

## auth.FindUserByEmail

Source: [internal/domain/auth/repository.go:34](../../../internal/domain/auth/repository.go#L34)

## auth.FindUserByUUID

Source: [internal/domain/auth/repository.go:18](../../../internal/domain/auth/repository.go#L18)

## auth.FindUserByUsername

Source: [internal/domain/auth/repository.go:14](../../../internal/domain/auth/repository.go#L14)

## auth.Message

Source: [internal/platform/mysql/auth/messages.go:20](../../../internal/platform/mysql/auth/messages.go#L20)

```sql
SELECT MESSAGE_DESCRIPTION_EN, MESSAGE_DESCRIPTION_TH FROM m_message_code WHERE MESSAGE_CODE = ? LIMIT 1
```

## auth.Register

Source: [internal/domain/auth/repository.go:38](../../../internal/domain/auth/repository.go#L38)

```sql
SELECT COUNT(1) FROM m_mobile_users WHERE EMAIL = ?
```

```sql
SELECT MAX(CAST(SMART_SEAMAN_ID AS DECIMAL) + 1) FROM m_mobile_users
```

```sql
INSERT INTO m_mobile_users
		(SMART_SEAMAN_ID, MOBILE_UUID, USERNAME, PASSWORD, FIRST_NAME, LAST_NAME, COMPANY_CODE, POSITION_CODE,
		 EMAIL, MOBILE_NUMBER, PROFILE_PICTURE, USER_STATUS, LAST_LOGON, CREATE_BY, CREATE_DATE, DATE_OF_BIRTH, DISPLAY_NAME, DISPLAY_TYPE)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'default', 'I', NULL, ?, NOW(), ?, ?, 'NAME')
```

```sql
SELECT DOCUMENT_CODE FROM m_documents WHERE DOCUMENT_DEFAULT_FLAG = 'Y' ORDER BY DOCUMENT_SEQ
```

```sql
INSERT INTO m_certificates
			(CERT_MOBILE_UUID, CERT_DOCUMENT_CODE, CERT_START_DATE, CERT_END_DATE, CERT_FILE, ORIGINAL_FILE_NAME,
			 CERT_STATUS, CREATE_DATE, CREATE_BY, UPDATE_DATE, UPDATE_BY)
			VALUES (?, ?, NULL, NULL, NULL, NULL, 'N', NOW(), ?, NOW(), ?)
```

## auth.RotateRefreshSession

Source: [internal/platform/mysql/auth/secure_repository.go:119](../../../internal/platform/mysql/auth/secure_repository.go#L119)

```sql
UPDATE mobile_auth_sessions a INNER JOIN t_session s ON s.CLIENT_SESSION_ID=a.session_id INNER JOIN m_mobile_users u ON u.MOBILE_UUID=s.USER_ID SET a.refresh_hash=?,s.TOKEN=? WHERE a.session_id=? AND a.refresh_hash=? AND a.expires_at>NOW() AND s.IS_ONLINE='YES' AND u.USER_STATUS='A'
```

## auth.UpdatePassword

Source: [internal/domain/auth/repository.go:100](../../../internal/domain/auth/repository.go#L100)

```sql
UPDATE m_mobile_users SET PASSWORD=? WHERE MOBILE_UUID=?
```

```sql
UPDATE t_session SET IS_ONLINE='NO' WHERE USER_ID=?
```

```sql
DELETE FROM mobile_auth_tokens WHERE user_id=? AND purpose='reset'
```

## auth.findUser

Source: [internal/domain/auth/repository.go:22](../../../internal/domain/auth/repository.go#L22)

```sql
SELECT MOBILE_UUID, USERNAME, PASSWORD, EMAIL, USER_STATUS, FIRST_NAME, LAST_NAME FROM m_mobile_users WHERE
```

