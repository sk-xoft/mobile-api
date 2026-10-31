# SQL catalog: notification

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## notification.Count

Source: [internal/domain/notification/repository.go:39](../../../internal/domain/notification/repository.go#L39)

```sql
SELECT COUNT(*) FROM m_send_notifications WHERE READ_STATUS='NO' AND MOBILE_USER_UUID=?
```

## notification.List

Source: [internal/domain/notification/repository.go:14](../../../internal/domain/notification/repository.go#L14)

```sql
SELECT id,READ_STATUS,NOTI_TYPE,MOBILE_USER_UUID,TITLE_MESSAGE,BODY_MESSAGE,SUCCESS,READ_DATE,RETRY,NOTI_DATE,VALUE_ID FROM m_send_notifications WHERE MOBILE_USER_UUID=? ORDER BY CREATE_DATE DESC
```

## notification.MarkAll

Source: [internal/domain/notification/repository.go:52](../../../internal/domain/notification/repository.go#L52)

```sql
UPDATE m_send_notifications SET READ_STATUS='YES',READ_DATE=NOW() WHERE MOBILE_USER_UUID=?
```

## notification.MarkID

Source: [internal/domain/notification/repository.go:44](../../../internal/domain/notification/repository.go#L44)

```sql
UPDATE m_send_notifications SET READ_STATUS='YES',READ_DATE=NOW() WHERE id=? AND MOBILE_USER_UUID=?
```

## notification.MarkValue

Source: [internal/domain/notification/repository.go:56](../../../internal/domain/notification/repository.go#L56)

```sql
UPDATE m_send_notifications SET READ_STATUS='YES',READ_DATE=NOW() WHERE MOBILE_USER_UUID=? AND VALUE_ID=? AND NOTI_TYPE=?
```

