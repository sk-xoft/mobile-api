# SQL catalog: fcm

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## fcm.Exists

Source: [internal/domain/fcm/repository.go:12](../../../internal/domain/fcm/repository.go#L12)

```sql
SELECT COUNT(*) FROM m_fcm_notifications WHERE USER_MOBILE_UUID=?
```

## fcm.Insert

Source: [internal/domain/fcm/repository.go:18](../../../internal/domain/fcm/repository.go#L18)

```sql
INSERT INTO m_fcm_notifications (USER_MOBILE_UUID, TOKEN_FCM, CREATE_DATE, UPDATE_DATE) VALUES (?, ?, NOW(), NOW())
```

## fcm.Update

Source: [internal/domain/fcm/repository.go:23](../../../internal/domain/fcm/repository.go#L23)

```sql
UPDATE m_fcm_notifications SET TOKEN_FCM=?, UPDATE_DATE=NOW() WHERE USER_MOBILE_UUID=?
```

