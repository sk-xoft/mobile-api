# SQL catalog: content

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## content.Form

Source: [internal/domain/content/repository.go:35](../../../internal/domain/content/repository.go#L35)

```sql
SELECT FORM_ID, FORM_NAME, FORM_FILE_ID, FORM_FILE_NAME FROM m_forms WHERE FORM_ID=?
```

## content.Forms

Source: [internal/domain/content/repository.go:16](../../../internal/domain/content/repository.go#L16)

```sql
SELECT FORM_ID, FORM_NAME, FORM_FILE_ID, FORM_FILE_NAME FROM m_forms
```

## content.News

Source: [internal/domain/content/repository.go:47](../../../internal/domain/content/repository.go#L47)

```sql
SELECT NEWS_ID, NEWS_TITLE, UPDATE_DATE FROM m_news WHERE NEWS_STATUS='A' AND NEWS_TYPE=? ORDER BY UPDATE_DATE DESC
```

## content.NewsByID

Source: [internal/domain/content/repository.go:70](../../../internal/domain/content/repository.go#L70)

```sql
SELECT NEWS_ID, NEWS_TITLE, NEWS_DETAILS, NEWS_PICTURE_FILE_NAME, UPDATE_DATE FROM m_news WHERE NEWS_ID=?
```

