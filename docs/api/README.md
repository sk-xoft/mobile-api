# API specifications

เอกสารภาษาไทยแยกหนึ่งไฟล์ต่อ HTTP method + route จาก `internal/transport/routes/routes.go`
อ้างอิง Go implementation ณ 2026-10-31; ไม่เปลี่ยน API behavior และไม่ execute SQL

- [Template สำหรับ endpoint ใหม่](TEMPLATE.md)
- [Headers, authentication, response และ common database access](COMMON.md)
- [JSON models ทุก domain](SCHEMAS.md)
- [ข้อจำกัดและความคลาดเคลื่อนที่พบ](KNOWN-GAPS.md)
- [OpenAPI ที่มีอยู่](../openapi.yaml): non-payment ยังเป็น route inventory; Markdown ชุดนี้อธิบาย business logic/query เพิ่มเติม

## วิธีดูแลเอกสาร

เมื่อแก้ route/handler/service/repository ให้แก้ไฟล์ endpoint, models และ SQL catalog ที่เกี่ยวข้องพร้อมกัน
ใช้ template และตรวจ required fields กับ validation จริง ไม่อนุมาน required จาก JSON tags
แยก domain SQL จาก middleware SQL และระบุ transaction/ownership/error mapping ตาม source
SQL catalog เป็น snippets สำหรับอ่านประกอบ ไม่ใช่ไฟล์นำไปรัน migration

## ขอบเขต

มี 75 method/routes ที่ register โดยตรง; CORS OPTIONS และ Fiber HEAD ที่ตาม GET ไม่แยกเป็น endpoint ใหม่
Swagger UI `/docs` และ `/docs/*` เปิดเฉพาะ API_DOCS_ENABLED=true ดู [COMMON.md](COMMON.md#optional-documentation-routes)
มี spec เพิ่มสำหรับ [GET /docs](system/get-docs.md) และ
[GET /docs/*](system/get-docs-resource.md) รวมเป็น 77 method/routes ในชุดเอกสารนี้
ไม่รวม method ที่มีใน source แต่ไม่ได้ bind route เช่น profile.Activate แบบ email-only

## auth

| Method | Path | Spec |
|---|---|---|
| POST | `/v1/login` | [เข้าสู่ระบบ](auth/post-v1-login.md) |
| POST | `/v1/register` | [สมัครสมาชิก](auth/post-v1-register.md) |
| GET | `/v1/refresh-token` | [หมุน refresh credential](auth/get-v1-refresh-token.md) |
| GET | `/v1/activate-user` | [ยืนยันบัญชีด้วย activation token](auth/get-v1-activate-user.md) |
| POST | `/v1/reset-password` | [ขอลิงก์ตั้งรหัสผ่านใหม่](auth/post-v1-reset-password.md) |
| POST | `/v1/activate-forgot-password` | [ตั้งรหัสผ่านจากลิงก์](auth/post-v1-activate-forgot-password.md) |
| POST | `/v1/profile/active` | [ยืนยันบัญชีด้วย activation token](auth/post-v1-profile-active.md) |
| POST | `/v1/change-password` | [เปลี่ยนรหัสผ่าน](auth/post-v1-change-password.md) |

## catalog

| Method | Path | Spec |
|---|---|---|
| GET | `/v1/banners` | [รายการแบนเนอร์](catalog/get-v1-banners.md) |
| GET | `/v1/banners/preview` | [ไฟล์แบนเนอร์](catalog/get-v1-banners-preview.md) |
| GET | `/v1/vouchers` | [คูปองที่แสดงแก่ผู้ใช้](catalog/get-v1-vouchers.md) |
| GET | `/v1/vouchers/detail` | [รายละเอียดคูปอง](catalog/get-v1-vouchers-detail.md) |
| GET | `/v1/vouchers/preview` | [รูปคูปอง](catalog/get-v1-vouchers-preview.md) |
| GET | `/v1/vouchers/qr` | [QR คูปอง](catalog/get-v1-vouchers-qr.md) |
| GET | `/v1/school-trainings` | [รายการโรงเรียนและหลักสูตร](catalog/get-v1-school-trainings.md) |
| GET | `/v1/school-trainings/detail` | [รอบอบรมของโรงเรียน](catalog/get-v1-school-trainings-detail.md) |

## content

| Method | Path | Spec |
|---|---|---|
| GET | `/v1/forms` | [รายการแบบฟอร์ม](content/get-v1-forms.md) |
| GET | `/v1/form` | [ดาวน์โหลดแบบฟอร์ม PDF](content/get-v1-form.md) |
| GET | `/v1/news` | [รายการข่าว](content/get-v1-news.md) |
| GET | `/v1/news/detail` | [รายละเอียดข่าว](content/get-v1-news-detail.md) |
| GET | `/v1/news/preview` | [รูปข่าว](content/get-v1-news-preview.md) |

## delivery

| Method | Path | Spec |
|---|---|---|
| GET | `/v1/delivery-addresses` | [ที่อยู่จัดส่งเริ่มต้น](delivery/get-v1-delivery-addresses.md) |
| POST | `/v1/delivery-addresses` | [สร้างที่อยู่จัดส่ง](delivery/post-v1-delivery-addresses.md) |
| PUT | `/v1/delivery-addresses/:addressId` | [แก้ไขที่อยู่จัดส่ง](delivery/put-v1-delivery-addresses-addressid.md) |
| POST | `/v1/delivery-addresses/:requestNo` | [สร้าง snapshot ที่อยู่ของคำขอ](delivery/post-v1-delivery-addresses-requestno.md) |

## document

| Method | Path | Spec |
|---|---|---|
| GET | `/v1/documents/certification/COT` | [รายการ certificate ตามประเภท](document/get-v1-documents-certification-cot.md) |
| GET | `/v1/documents/certification/DOC` | [รายการ certificate ตามประเภท](document/get-v1-documents-certification-doc.md) |
| GET | `/v1/documents/certification/to-expiration` | [certificate ที่ใกล้หมดอายุ](document/get-v1-documents-certification-to-expiration.md) |
| POST | `/v1/documents/certification/create` | [เพิ่ม certificate](document/post-v1-documents-certification-create.md) |
| POST | `/v1/documents/certification/update` | [แก้ไข certificate](document/post-v1-documents-certification-update.md) |
| DELETE | `/v1/documents/certification/delete` | [ลบ certificate](document/delete-v1-documents-certification-delete.md) |
| GET | `/v1/documents/certification/edit` | [ข้อมูล certificate สำหรับแก้ไข](document/get-v1-documents-certification-edit.md) |
| GET | `/v1/documents/certification/view` | [ไฟล์ certificate](document/get-v1-documents-certification-view.md) |

## fcm

| Method | Path | Spec |
|---|---|---|
| POST | `/v1/fcm-update` | [บันทึก FCM token](fcm/post-v1-fcm-update.md) |

## master

| Method | Path | Spec |
|---|---|---|
| GET | `/v1/master` | [ข้อมูลบริษัทและตำแหน่ง](master/get-v1-master.md) |
| GET | `/v1/master/documents` | [รายการประเภทเอกสารและ COT](master/get-v1-master-documents.md) |
| GET | `/v1/master/courses` | [รายชื่อหลักสูตร](master/get-v1-master-courses.md) |
| GET | `/v1/master/provinces` | [รายชื่อจังหวัด](master/get-v1-master-provinces.md) |
| GET | `/v1/master/districts` | [รายชื่ออำเภอ](master/get-v1-master-districts.md) |
| GET | `/v1/master/subdistricts` | [รายชื่อตำบล](master/get-v1-master-subdistricts.md) |

## notification

| Method | Path | Spec |
|---|---|---|
| GET | `/v1/send-noti` | [ตอบรับ manual notification route](notification/get-v1-send-noti.md) |
| GET | `/v1/notifications` | [รายการแจ้งเตือนของฉัน](notification/get-v1-notifications.md) |
| POST | `/v1/notifications/update` | [อ่าน notification ตาม ID](notification/post-v1-notifications-update.md) |
| POST | `/v1/notifications/update/valueId` | [อ่าน notification ตาม value/type](notification/post-v1-notifications-update-valueid.md) |
| GET | `/v1/notifications/update/all` | [อ่าน notification ทั้งหมด](notification/get-v1-notifications-update-all.md) |

## profile

| Method | Path | Spec |
|---|---|---|
| GET | `/v1/profile` | [อ่านโปรไฟล์](profile/get-v1-profile.md) |
| POST | `/v1/profile-update` | [แก้ไขโปรไฟล์](profile/post-v1-profile-update.md) |
| GET | `/v1/profile-image` | [อ่านรูปโปรไฟล์](profile/get-v1-profile-image.md) |
| DELETE | `/v1/profile/inactive` | [ปิดบัญชี](profile/delete-v1-profile-inactive.md) |

## renewal

| Method | Path | Spec |
|---|---|---|
| POST | `/v1/payments/omise/webhook` | [รับ Omise webhook](renewal/post-v1-payments-omise-webhook.md) |
| GET | `/v1/document-renewals/statuses` | [สถานะต่ออายุเอกสาร](renewal/get-v1-document-renewals-statuses.md) |
| GET | `/v1/document-renewals/prices` | [ราคาต่ออายุเอกสาร](renewal/get-v1-document-renewals-prices.md) |
| GET | `/v1/document-renewals` | [เอกสารที่ต่ออายุได้](renewal/get-v1-document-renewals.md) |
| POST | `/v1/document-renewals` | [สร้างคำขอต่ออายุแบบเดิม](renewal/post-v1-document-renewals.md) |
| POST | `/v1/document-renewals/requests/validate-and-create` | [ตรวจและสร้าง unpaid draft](renewal/post-v1-document-renewals-requests-validate-and-create.md) |
| GET | `/v1/document-renewals/my` | [รายการคำขอต่ออายุของฉัน](renewal/get-v1-document-renewals-my.md) |
| PUT | `/v1/document-renewals/requests/:requestNo/mobile` | [แก้เบอร์ snapshot ของคำขอ](renewal/put-v1-document-renewals-requests-requestno-mobile.md) |
| DELETE | `/v1/document-renewals/requests/:requestNo` | [ยกเลิกคำขอแบบ soft delete](renewal/delete-v1-document-renewals-requests-requestno.md) |
| DELETE | `/v1/document-renewals/requests/:requestNo/hard` | [ลบคำขอและข้อมูลลูกถาวร](renewal/delete-v1-document-renewals-requests-requestno-hard.md) |
| POST | `/v1/document-renewals/:requestNo/resubmit` | [ส่งเอกสารที่แก้ไขแล้วอีกครั้ง](renewal/post-v1-document-renewals-requestno-resubmit.md) |
| GET | `/v1/document-renewals/:requestNo/timeline` | [ประวัติขั้นตอนคำขอ](renewal/get-v1-document-renewals-requestno-timeline.md) |
| GET | `/v1/document-renewals/:requestNo/stage` | [ความคืบหน้า 5 ขั้นตอน](renewal/get-v1-document-renewals-requestno-stage.md) |
| GET | `/v1/document-renewals/:requestNo` | [รายละเอียดคำขอต่ออายุ](renewal/get-v1-document-renewals-requestno.md) |
| POST | `/v1/document-renewals/:requestId/payments` | [สร้างหรือ replay payment attempt](renewal/post-v1-document-renewals-requestid-payments.md) |
| GET | `/v1/document-renewals/:requestId/payments` | [อ่านและ reconcile payment](renewal/get-v1-document-renewals-requestid-payments.md) |
| GET | `/v1/document-renewals/:requestId/payments/:transactionId` | [อ่านและ reconcile payment](renewal/get-v1-document-renewals-requestid-payments-transactionid.md) |

## requestitem

| Method | Path | Spec |
|---|---|---|
| POST | `/v1/documents/request-items/:itemCode/files` | [อัปโหลดไฟล์เอกสารโปรไฟล์](requestitem/post-v1-documents-request-items-itemcode-files.md) |
| GET | `/v1/documents/profile-items` | [รายการเอกสารโปรไฟล์และไฟล์](requestitem/get-v1-documents-profile-items.md) |
| GET | `/v1/documents/profile-items/:itemCode/preview` | [metadata ไฟล์เอกสารโปรไฟล์](requestitem/get-v1-documents-profile-items-itemcode-preview.md) |
| GET | `/v1/document-renewals/items/preview` | [metadata ไฟล์ของคำขอต่ออายุ](requestitem/get-v1-document-renewals-items-preview.md) |
| GET | `/v1/document-renewals/:requestNo/items/:documentRequestItemCode/preview` | [metadata ไฟล์ของคำขอต่ออายุ](requestitem/get-v1-document-renewals-requestno-items-documentrequestitemcode-preview.md) |
| POST | `/v1/document-renewals/:requestNo/items/:documentRequestItemCode/file` | [อัปโหลดไฟล์ของคำขอต่ออายุ](requestitem/post-v1-document-renewals-requestno-items-documentrequestitemcode-file.md) |

## system

| Method | Path | Spec |
|---|---|---|
| GET | `/` | [ข้อความต้อนรับ](system/get-.md) |
| GET | `/health` | [health probe](system/get-health.md) |
| GET | `/privacy_policy` | [privacy policy placeholder](system/get-privacy-policy.md) |

