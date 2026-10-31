# SQL catalog: profile

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## profile.Descriptions

Source: [internal/domain/profile/repository.go:22](../../../internal/domain/profile/repository.go#L22)

```sql
SELECT
```

## profile.FindUser

Source: [internal/domain/profile/repository.go:12](../../../internal/domain/profile/repository.go#L12)

```sql
SELECT MOBILE_UUID, SMART_SEAMAN_ID, USERNAME, EMAIL, FIRST_NAME, LAST_NAME,
		MOBILE_NUMBER, COMPANY_CODE, POSITION_CODE, PROFILE_PICTURE, USER_STATUS, DATE_OF_BIRTH
		FROM m_mobile_users WHERE MOBILE_UUID = ? LIMIT 1
```

## profile.Update

Source: [internal/domain/profile/repository.go:35](../../../internal/domain/profile/repository.go#L35)

```sql
SELECT MOBILE_NUMBER FROM m_mobile_users WHERE MOBILE_UUID = ? FOR UPDATE
```

```sql
INSERT INTO m_mobile_number_history (id, mobile_user_uuid, old_mobile_number, new_mobile_number, changed_by) VALUES (UUID(), ?, ?, ?, ?)
```

```sql
UPDATE m_mobile_users SET FIRST_NAME=?, LAST_NAME=?, COMPANY_CODE=?, POSITION_CODE=?, MOBILE_NUMBER=?, DATE_OF_BIRTH=?, UPDATE_DATE=NOW(), UPDATE_BY=? WHERE MOBILE_UUID=?
```

```sql
UPDATE m_mobile_users SET PROFILE_PICTURE=?, UPDATE_DATE=NOW(), UPDATE_BY=? WHERE MOBILE_UUID=?
```

## profile.UpdateStatus

Source: [internal/domain/profile/repository.go:62](../../../internal/domain/profile/repository.go#L62)

```sql
SELECT MOBILE_UUID, SMART_SEAMAN_ID, USERNAME, EMAIL, FIRST_NAME, LAST_NAME, MOBILE_NUMBER, COMPANY_CODE, POSITION_CODE, PROFILE_PICTURE, USER_STATUS, DATE_OF_BIRTH FROM m_mobile_users WHERE EMAIL=? LIMIT 1
```

```sql
UPDATE m_mobile_users SET USER_STATUS=?, UPDATE_DATE=NOW(), UPDATE_BY=? WHERE MOBILE_UUID=?
```

```sql
UPDATE t_session SET IS_ONLINE='NO' WHERE USER_ID=?
```

```sql
DELETE FROM mobile_auth_tokens WHERE user_id=?
```

