# JSON models

ชื่อ field และชนิดจาก Go JSON tags ณ 2026-10-31 ใช้ร่วมกับกฎ validation ในแต่ละ endpoint
`omitempty` หมายถึงอาจไม่ส่ง field เมื่อเป็น zero value ไม่ได้แปลว่า input optional
`*T` หมายถึงอาจเป็น null; `[]T` เป็น array; Go string ยังคงเป็น JSON string แม้มีข้อมูลตัวเลข
ตัวอย่างใน endpoint เป็น structural illustration ไม่ใช่ผลทดสอบจริง; field nested ดู model ที่ลิงก์

## auth.ActivationInput

Source: [internal/domain/auth/handler.go](../../internal/domain/auth/handler.go)

| JSON field | Type | Serialization |
|---|---|---|
| `uid` | `string` | ส่ง field; pointer อาจเป็น null |

## auth.ChangePasswordRequest

Source: [internal/domain/auth/model.go](../../internal/domain/auth/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `oldPassword` | `string` | ส่ง field; pointer อาจเป็น null |
| `newPassword` | `string` | ส่ง field; pointer อาจเป็น null |
| `confirmPassword` | `string` | ส่ง field; pointer อาจเป็น null |

## auth.ForgotPasswordRequest

Source: [internal/domain/auth/model.go](../../internal/domain/auth/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `password` | `string` | ส่ง field; pointer อาจเป็น null |
| `confirmPassword` | `string` | ส่ง field; pointer อาจเป็น null |
| `uid` | `string` | ส่ง field; pointer อาจเป็น null |

## auth.LoginRequest

Source: [internal/domain/auth/model.go](../../internal/domain/auth/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `username` | `string` | ส่ง field; pointer อาจเป็น null |
| `password` | `string` | ส่ง field; pointer อาจเป็น null |

## auth.LoginResponse

Source: [internal/domain/auth/model.go](../../internal/domain/auth/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `token` | `string` | ส่ง field; pointer อาจเป็น null |
| `refToken` | `string` | ส่ง field; pointer อาจเป็น null |
| `username` | `string` | ส่ง field; pointer อาจเป็น null |
| `lastLoginDateTime` | `string` | ส่ง field; pointer อาจเป็น null |

## auth.RefreshTokenResponse

Source: [internal/domain/auth/model.go](../../internal/domain/auth/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `refToken` | `string` | ส่ง field; pointer อาจเป็น null |
| `token` | `string` | ส่ง field; pointer อาจเป็น null |

## auth.RegisterRequest

Source: [internal/domain/auth/model.go](../../internal/domain/auth/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `firstName` | `string` | ส่ง field; pointer อาจเป็น null |
| `lastName` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `dateOfBirth` | `string` | ส่ง field; pointer อาจเป็น null |
| `positionCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `email` | `string` | ส่ง field; pointer อาจเป็น null |
| `password` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobileNumber` | `string` | ส่ง field; pointer อาจเป็น null |

## auth.RegisterResponse

Source: [internal/domain/auth/model.go](../../internal/domain/auth/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `username` | `string` | ส่ง field; pointer อาจเป็น null |
| `email` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobileNumber` | `string` | อาจไม่ส่งเมื่อ zero value |

## auth.ResetPasswordRequest

Source: [internal/domain/auth/model.go](../../internal/domain/auth/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `email` | `string` | ส่ง field; pointer อาจเป็น null |

## catalog.Banner

Source: [internal/domain/catalog/model.go](../../internal/domain/catalog/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `bannerId` | `string` | ส่ง field; pointer อาจเป็น null |
| `bannerFileName` | `string` | ส่ง field; pointer อาจเป็น null |

## catalog.BannerResponse

Source: [internal/domain/catalog/model.go](../../internal/domain/catalog/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `banners` | array of [catalog.Banner](SCHEMAS.md#catalogbanner) | ส่ง field; pointer อาจเป็น null |

## catalog.School

Source: [internal/domain/catalog/model.go](../../internal/domain/catalog/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `companyCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyNameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyNameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyColour` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseColour` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseId` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseNameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseNameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseType` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseOnlineDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseOnsiteDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseTotalDays` | `string` | ส่ง field; pointer อาจเป็น null |
| `coursePrice` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyLine` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyFacebook` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyPhone1` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseStartDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseEndDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `dateForCheck` | `string` | ส่ง field; pointer อาจเป็น null |

## catalog.SchoolResponse

Source: [internal/domain/catalog/model.go](../../internal/domain/catalog/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `schoolTrainings` | array of [catalog.School](SCHEMAS.md#catalogschool) | ส่ง field; pointer อาจเป็น null |

## catalog.Voucher

Source: [internal/domain/catalog/model.go](../../internal/domain/catalog/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `voucherId` | `string` | ส่ง field; pointer อาจเป็น null |
| `voucherTitle` | `string` | ส่ง field; pointer อาจเป็น null |
| `voucherDetails` | `string` | ส่ง field; pointer อาจเป็น null |
| `voucherStartDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `voucherEndDate` | `string` | ส่ง field; pointer อาจเป็น null |

## catalog.VoucherResponse

Source: [internal/domain/catalog/model.go](../../internal/domain/catalog/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `voucherModel` | array of [catalog.Voucher](SCHEMAS.md#catalogvoucher) | ส่ง field; pointer อาจเป็น null |

## content.Form

Source: [internal/domain/content/model.go](../../internal/domain/content/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `formId` | `string` | ส่ง field; pointer อาจเป็น null |
| `formName` | `string` | ส่ง field; pointer อาจเป็น null |
| `formFileId` | `string` | ส่ง field; pointer อาจเป็น null |
| `formFileName` | `string` | ส่ง field; pointer อาจเป็น null |

## content.FormsResponse

Source: [internal/domain/content/model.go](../../internal/domain/content/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `forms` | array of [content.Form](SCHEMAS.md#contentform) | ส่ง field; pointer อาจเป็น null |

## content.News

Source: [internal/domain/content/model.go](../../internal/domain/content/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `title` | `string` | ส่ง field; pointer อาจเป็น null |
| `body` | `string` | ส่ง field; pointer อาจเป็น null |
| `newsDate` | `string` | ส่ง field; pointer อาจเป็น null |

## content.NewsResponse

Source: [internal/domain/content/model.go](../../internal/domain/content/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `newsGeneral` | array of [content.News](SCHEMAS.md#contentnews) | ส่ง field; pointer อาจเป็น null |
| `newsShip` | array of [content.News](SCHEMAS.md#contentnews) | ส่ง field; pointer อาจเป็น null |

## delivery.Request

Source: [internal/domain/delivery/model.go](../../internal/domain/delivery/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `firstName` | `string` | ส่ง field; pointer อาจเป็น null |
| `lastName` | `string` | ส่ง field; pointer อาจเป็น null |
| `addressLine` | `string` | ส่ง field; pointer อาจเป็น null |
| `province` | `string` | ส่ง field; pointer อาจเป็น null |
| `district` | `string` | ส่ง field; pointer อาจเป็น null |
| `subDistrict` | `string` | ส่ง field; pointer อาจเป็น null |
| `postalCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `isDefault` | `*bool` | ส่ง field; pointer อาจเป็น null |

## delivery.Response

Source: [internal/domain/delivery/model.go](../../internal/domain/delivery/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `firstName` | `string` | ส่ง field; pointer อาจเป็น null |
| `lastName` | `string` | ส่ง field; pointer อาจเป็น null |
| `addressLine` | `string` | ส่ง field; pointer อาจเป็น null |
| `province` | `string` | ส่ง field; pointer อาจเป็น null |
| `district` | `string` | ส่ง field; pointer อาจเป็น null |
| `subDistrict` | `string` | ส่ง field; pointer อาจเป็น null |
| `postalCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobileNumber` | `string` | ส่ง field; pointer อาจเป็น null |
| `description` | `string` | ส่ง field; pointer อาจเป็น null |
| `isDefault` | `bool` | ส่ง field; pointer อาจเป็น null |

## document.CreateRequest

Source: [internal/domain/document/model.go](../../internal/domain/document/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `certStartDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `certEndDateType` | `string` | ส่ง field; pointer อาจเป็น null |
| `certEndDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `fileCert` | `string` | ส่ง field; pointer อาจเป็น null |
| `fileCertName` | `string` | ส่ง field; pointer อาจเป็น null |

## document.Item

Source: [internal/domain/document/model.go](../../internal/domain/document/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentId` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentNameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentNameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentDescription` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentType` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentDefaultFlag` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentRenewalFlag` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentSeq` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentMobileFlag` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentCourseCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `certStartDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `certEndDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `certFileName` | `string` | ส่ง field; pointer อาจเป็น null |
| `disYear` | `string` | ส่ง field; pointer อาจเป็น null |
| `disMonth` | `string` | ส่ง field; pointer อาจเป็น null |
| `disDay` | `string` | ส่ง field; pointer อาจเป็น null |
| `certExpiredFlag` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentRenewalRequestFlag` | `string` | ส่ง field; pointer อาจเป็น null |

## document.PageResponse

Source: [internal/domain/document/model.go](../../internal/domain/document/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `itemTotal` | `int` | ส่ง field; pointer อาจเป็น null |
| `last` | `bool` | ส่ง field; pointer อาจเป็น null |
| `items` | array of [document.Item](SCHEMAS.md#documentitem) | ส่ง field; pointer อาจเป็น null |

## document.Response

Source: [internal/domain/document/model.go](../../internal/domain/document/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `certStartDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `certEndDateType` | `string` | ส่ง field; pointer อาจเป็น null |
| `certEndDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `fileCertName` | `string` | ส่ง field; pointer อาจเป็น null |

## document.UpdateRequest

Source: [internal/domain/document/model.go](../../internal/domain/document/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `isChangeFile` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `certStartDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `certEndDateType` | `string` | ส่ง field; pointer อาจเป็น null |
| `certEndDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `fileCert` | `string` | ส่ง field; pointer อาจเป็น null |
| `fileCertName` | `string` | ส่ง field; pointer อาจเป็น null |

## fcm.UpdateRequest

Source: [internal/domain/fcm/model.go](../../internal/domain/fcm/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `tokenFcm` | `string` | ส่ง field; pointer อาจเป็น null |

## master.Address

Source: [internal/domain/master/model.go](../../internal/domain/master/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `code` | `string` | ส่ง field; pointer อาจเป็น null |
| `name` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `postalCode` | `string` | อาจไม่ส่งเมื่อ zero value |

## master.Company

Source: [internal/domain/master/model.go](../../internal/domain/master/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `companyCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyName` | `string` | ส่ง field; pointer อาจเป็น null |

## master.Course

Source: [internal/domain/master/model.go](../../internal/domain/master/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `courseCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseNameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `courseNameTh` | `string` | ส่ง field; pointer อาจเป็น null |

## master.CoursesResponse

Source: [internal/domain/master/model.go](../../internal/domain/master/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `courses` | array of [master.Course](SCHEMAS.md#mastercourse) | ส่ง field; pointer อาจเป็น null |

## master.Document

Source: [internal/domain/master/model.go](../../internal/domain/master/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentName` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentNameTh` | `string` | ส่ง field; pointer อาจเป็น null |

## master.DocumentsResponse

Source: [internal/domain/master/model.go](../../internal/domain/master/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documents` | array of [master.Document](SCHEMAS.md#masterdocument) | ส่ง field; pointer อาจเป็น null |
| `cot` | array of [master.Document](SCHEMAS.md#masterdocument) | ส่ง field; pointer อาจเป็น null |

## master.MasterResponse

Source: [internal/domain/master/model.go](../../internal/domain/master/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `company` | array of [master.Company](SCHEMAS.md#mastercompany) | ส่ง field; pointer อาจเป็น null |
| `position` | array of [master.Position](SCHEMAS.md#masterposition) | ส่ง field; pointer อาจเป็น null |

## master.Position

Source: [internal/domain/master/model.go](../../internal/domain/master/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `positionCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `positionName` | `string` | ส่ง field; pointer อาจเป็น null |

## notification.CountResponse

Source: [internal/domain/notification/model.go](../../internal/domain/notification/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `countNotification` | `string` | ส่ง field; pointer อาจเป็น null |

## notification.Item

Source: [internal/domain/notification/model.go](../../internal/domain/notification/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `readStatus` | `string` | ส่ง field; pointer อาจเป็น null |
| `notiType` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobileUserUuid` | `string` | ส่ง field; pointer อาจเป็น null |
| `titleMessage` | `string` | ส่ง field; pointer อาจเป็น null |
| `bodyMessage` | `string` | ส่ง field; pointer อาจเป็น null |
| `success` | `string` | ส่ง field; pointer อาจเป็น null |
| `readDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `retry` | `string` | ส่ง field; pointer อาจเป็น null |
| `notiDate` | `string` | ส่ง field; pointer อาจเป็น null |
| `valueId` | `string` | ส่ง field; pointer อาจเป็น null |

## notification.MarkRequest

Source: [internal/domain/notification/model.go](../../internal/domain/notification/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `notiId` | `string` | ส่ง field; pointer อาจเป็น null |

## notification.MarkValueRequest

Source: [internal/domain/notification/model.go](../../internal/domain/notification/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `valueId` | `string` | ส่ง field; pointer อาจเป็น null |
| `notiType` | `string` | ส่ง field; pointer อาจเป็น null |

## notification.Response

Source: [internal/domain/notification/model.go](../../internal/domain/notification/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `notificationsModels` | array of [notification.Item](SCHEMAS.md#notificationitem) | ส่ง field; pointer อาจเป็น null |

## payment.Charge

Source: [internal/domain/payment/model.go](../../internal/domain/payment/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `ref_id` | `string` | ส่ง field; pointer อาจเป็น null |
| `request_no` | `string` | ส่ง field; pointer อาจเป็น null |
| `status` | `string` | ส่ง field; pointer อาจเป็น null |
| `amount` | `int64` | ส่ง field; pointer อาจเป็น null |
| `currency` | `string` | ส่ง field; pointer อาจเป็น null |
| `livemode` | `bool` | ส่ง field; pointer อาจเป็น null |
| `expired` | `bool` | ส่ง field; pointer อาจเป็น null |
| `expires_at` | `time.Time` | ส่ง field; pointer อาจเป็น null |

## payment.Input

Source: [internal/domain/payment/model.go](../../internal/domain/payment/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `paymentMethod` | `string` | ส่ง field; pointer อาจเป็น null |
| `idempotencyKey` | `string` | ส่ง field; pointer อาจเป็น null |
| `returnUri` | `string` | ส่ง field; pointer อาจเป็น null |
| `platformType` | `string` | ส่ง field; pointer อาจเป็น null |

## payment.Response

Source: [internal/domain/payment/model.go](../../internal/domain/payment/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestId` | `string` | ส่ง field; pointer อาจเป็น null |
| `requestNo` | `string` | ส่ง field; pointer อาจเป็น null |
| `transactionId` | `string` | ส่ง field; pointer อาจเป็น null |
| `transactionNo` | `string` | ส่ง field; pointer อาจเป็น null |
| `status` | `string` | ส่ง field; pointer อาจเป็น null |
| `channel` | `string` | ส่ง field; pointer อาจเป็น null |
| `paymentMethod` | `string` | ส่ง field; pointer อาจเป็น null |
| `amount` | `float64` | ส่ง field; pointer อาจเป็น null |
| `currency` | `string` | ส่ง field; pointer อาจเป็น null |
| `provider` | `string` | ส่ง field; pointer อาจเป็น null |
| `providerChargeId` | `string` | อาจไม่ส่งเมื่อ zero value |
| `authorizeUri` | `string` | อาจไม่ส่งเมื่อ zero value |
| `returnUri` | `string` | อาจไม่ส่งเมื่อ zero value |
| `qrCodeDownloadUri` | `string` | อาจไม่ส่งเมื่อ zero value |
| `expiresAt` | `string` | อาจไม่ส่งเมื่อ zero value |
| `failureCode` | `string` | อาจไม่ส่งเมื่อ zero value |
| `failureMessage` | `string` | อาจไม่ส่งเมื่อ zero value |
| `serverTime` | `string` | ส่ง field; pointer อาจเป็น null |
| `qrExpired` | `bool` | ส่ง field; pointer อาจเป็น null |
| `nextAction` | `string` | ส่ง field; pointer อาจเป็น null |
| `retryAfterSeconds` | `int` | ส่ง field; pointer อาจเป็น null |

## payment.WebhookData

Source: [internal/domain/payment/service.go](../../internal/domain/payment/service.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |

## payment.WebhookInput

Source: [internal/domain/payment/service.go](../../internal/domain/payment/service.go)

| JSON field | Type | Serialization |
|---|---|---|
| `key` | `string` | ส่ง field; pointer อาจเป็น null |
| `data` | [payment.WebhookData](SCHEMAS.md#paymentwebhookdata) | ส่ง field; pointer อาจเป็น null |

## profile.ActivateRequest

Source: [internal/domain/profile/model.go](../../internal/domain/profile/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `email` | `string` | ส่ง field; pointer อาจเป็น null |

## profile.RegisterResponse

Source: [internal/domain/profile/model.go](../../internal/domain/profile/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `username` | `string` | ส่ง field; pointer อาจเป็น null |
| `email` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobileNumber` | `string` | อาจไม่ส่งเมื่อ zero value |

## profile.Response

Source: [internal/domain/profile/model.go](../../internal/domain/profile/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `firstName` | `string` | ส่ง field; pointer อาจเป็น null |
| `lastName` | `string` | ส่ง field; pointer อาจเป็น null |
| `dateOfBirth` | `string` | ส่ง field; pointer อาจเป็น null |
| `age` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobile` | `string` | ส่ง field; pointer อาจเป็น null |
| `email` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyDescription` | `string` | ส่ง field; pointer อาจเป็น null |
| `positionCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `positionDescription` | `string` | ส่ง field; pointer อาจเป็น null |
| `smartSeamanId` | `string` | ส่ง field; pointer อาจเป็น null |
| `shortName` | `string` | ส่ง field; pointer อาจเป็น null |

## profile.UpdateRequest

Source: [internal/domain/profile/model.go](../../internal/domain/profile/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `firstName` | `string` | ส่ง field; pointer อาจเป็น null |
| `lastName` | `string` | ส่ง field; pointer อาจเป็น null |
| `companyCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `dateOfBirth` | `string` | ส่ง field; pointer อาจเป็น null |
| `positionCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `email` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobileNumber` | `string` | ส่ง field; pointer อาจเป็น null |
| `isChangeFile` | `string` | ส่ง field; pointer อาจเป็น null |
| `imageProfile` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.CreateRequest

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `deliveryAddressId` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.CreateResponse

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestId` | `string` | ส่ง field; pointer อาจเป็น null |
| `requestNo` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `status` | `string` | ส่ง field; pointer อาจเป็น null |
| `amount` | `float64` | ส่ง field; pointer อาจเป็น null |
| `deliveryAddressId` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.Detail

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestId` | `string` | ส่ง field; pointer อาจเป็น null |
| `requestNo` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobileUserUuid` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `status` | `string` | ส่ง field; pointer อาจเป็น null |
| `amount` | `float64` | ส่ง field; pointer อาจเป็น null |
| `submittedAt` | `string` | ส่ง field; pointer อาจเป็น null |
| `isResubmit` | `bool` | ส่ง field; pointer อาจเป็น null |

## renewal.Document

Source: [internal/domain/renewal/model.go](../../internal/domain/renewal/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentName` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentNameTh` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.MobileRequest

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `mobileNumber` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.MobileStatus

Source: [internal/domain/renewal/model.go](../../internal/domain/renewal/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentMobileStatusCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `step` | `int` | ส่ง field; pointer อาจเป็น null |

## renewal.Page

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `itemTotal` | `int` | ส่ง field; pointer อาจเป็น null |
| `last` | `bool` | ส่ง field; pointer อาจเป็น null |
| `items` | array of [renewal.Summary](SCHEMAS.md#renewalsummary) | ส่ง field; pointer อาจเป็น null |

## renewal.Price

Source: [internal/domain/renewal/model.go](../../internal/domain/renewal/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `governmentFee` | `float64` | ส่ง field; pointer อาจเป็น null |
| `documentProcessingFee` | `float64` | ส่ง field; pointer อาจเป็น null |
| `shippingFee` | `float64` | ส่ง field; pointer อาจเป็น null |
| `shippingDiscount` | `float64` | ส่ง field; pointer อาจเป็น null |
| `serviceFeeDiscount` | `float64` | ส่ง field; pointer อาจเป็น null |
| `total` | `float64` | ส่ง field; pointer อาจเป็น null |
| `effectiveFrom` | `*time.Time` | ส่ง field; pointer อาจเป็น null |
| `effectiveTo` | `*time.Time` | ส่ง field; pointer อาจเป็น null |

## renewal.Stage

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `step` | `int` | ส่ง field; pointer อาจเป็น null |
| `documentMobileStatusCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `state` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.StageCurrentStatus

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentStatusCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `cssColor` | `string` | ส่ง field; pointer อาจเป็น null |
| `step` | `*int` | ส่ง field; pointer อาจเป็น null |
| `mobileStatus` | nullable [renewal.StageMobileStatus](SCHEMAS.md#renewalstagemobilestatus) | ส่ง field; pointer อาจเป็น null |

## renewal.StageMobileStatus

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentMobileStatusCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `step` | `int` | ส่ง field; pointer อาจเป็น null |

## renewal.StageResponse

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestNo` | `string` | ส่ง field; pointer อาจเป็น null |
| `currentStatus` | [renewal.StageCurrentStatus](SCHEMAS.md#renewalstagecurrentstatus) | ส่ง field; pointer อาจเป็น null |
| `stages` | array of [renewal.Stage](SCHEMAS.md#renewalstage) | ส่ง field; pointer อาจเป็น null |

## renewal.Status

Source: [internal/domain/renewal/model.go](../../internal/domain/renewal/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentStatusCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameTh` | `string` | ส่ง field; pointer อาจเป็น null |
| `nameEn` | `string` | ส่ง field; pointer อาจเป็น null |
| `cssColor` | `string` | ส่ง field; pointer อาจเป็น null |
| `progressStep` | `*int` | ส่ง field; pointer อาจเป็น null |
| `mobileStatus` | nullable [renewal.MobileStatus](SCHEMAS.md#renewalmobilestatus) | ส่ง field; pointer อาจเป็น null |
| `correction` | `bool` | ส่ง field; pointer อาจเป็น null |
| `terminal` | `bool` | ส่ง field; pointer อาจเป็น null |

## renewal.Summary

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestId` | `string` | ส่ง field; pointer อาจเป็น null |
| `requestNo` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `status` | `string` | ส่ง field; pointer อาจเป็น null |
| `statusName` | `string` | ส่ง field; pointer อาจเป็น null |
| `amount` | `float64` | ส่ง field; pointer อาจเป็น null |
| `submittedAt` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.Timeline

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestNo` | `string` | ส่ง field; pointer อาจเป็น null |
| `items` | array of [renewal.TimelineItem](SCHEMAS.md#renewaltimelineitem) | ส่ง field; pointer อาจเป็น null |

## renewal.TimelineItem

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `action` | `string` | ส่ง field; pointer อาจเป็น null |
| `fromStatus` | `string` | ส่ง field; pointer อาจเป็น null |
| `toStatus` | `string` | ส่ง field; pointer อาจเป็น null |
| `note` | `string` | ส่ง field; pointer อาจเป็น null |
| `actionedAt` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.TransitionResponse

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestNo` | `string` | ส่ง field; pointer อาจเป็น null |
| `fromStatus` | `string` | ส่ง field; pointer อาจเป็น null |
| `toStatus` | `string` | ส่ง field; pointer อาจเป็น null |
| `action` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.ValidateAddress

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `firstName` | `string` | ส่ง field; pointer อาจเป็น null |
| `lastName` | `string` | ส่ง field; pointer อาจเป็น null |
| `addressLine` | `string` | ส่ง field; pointer อาจเป็น null |
| `province` | `string` | ส่ง field; pointer อาจเป็น null |
| `district` | `string` | ส่ง field; pointer อาจเป็น null |
| `subDistrict` | `string` | ส่ง field; pointer อาจเป็น null |
| `postalCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `mobileNumber` | `string` | ส่ง field; pointer อาจเป็น null |
| `description` | `string` | อาจไม่ส่งเมื่อ zero value |
| `isDefault` | `*bool` | ส่ง field; pointer อาจเป็น null |

## renewal.ValidateFile

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `fileId` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentType` | `string` | ส่ง field; pointer อาจเป็น null |
| `slotCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `filePath` | `string` | อาจไม่ส่งเมื่อ zero value |
| `originalFileName` | `string` | ส่ง field; pointer อาจเป็น null |
| `mimeType` | `string` | ส่ง field; pointer อาจเป็น null |
| `fileSize` | `int64` | ส่ง field; pointer อาจเป็น null |
| `fileUploadedAt` | `string` | อาจไม่ส่งเมื่อ zero value |
| `checkResult` | `string` | อาจไม่ส่งเมื่อ zero value |
| `checkNote` | `string` | อาจไม่ส่งเมื่อ zero value |
| `isUpdated` | `bool` | อาจไม่ส่งเมื่อ zero value |

## renewal.ValidateInput

Source: [internal/domain/renewal/lifecycle_handler.go](../../internal/domain/renewal/lifecycle_handler.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `idempotencyKey` | `string` | ส่ง field; pointer อาจเป็น null |

## renewal.ValidateItem

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `profileRequestItemId` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentMasterRequestItemCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `storageScope` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentType` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentName` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentStatus` | `string` | ส่ง field; pointer อาจเป็น null |
| `sortOrder` | `int` | ส่ง field; pointer อาจเป็น null |
| `fileUploaded` | `int` | ส่ง field; pointer อาจเป็น null |
| `filePath` | `string` | อาจไม่ส่งเมื่อ zero value |
| `fileUploadedAt` | `string` | อาจไม่ส่งเมื่อ zero value |
| `checkResult` | `string` | อาจไม่ส่งเมื่อ zero value |
| `checkNote` | `string` | อาจไม่ส่งเมื่อ zero value |
| `files` | array of [renewal.ValidateFile](SCHEMAS.md#renewalvalidatefile) | อาจไม่ส่งเมื่อ zero value |

## renewal.ValidateResponse

Source: [internal/domain/renewal/lifecycle.go](../../internal/domain/renewal/lifecycle.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestId` | `string` | อาจไม่ส่งเมื่อ zero value |
| `requestNo` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentName` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentNameTh` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentNameEn` | `string` | อาจไม่ส่งเมื่อ zero value |
| `idempotencyKey` | `*string` | ส่ง field; pointer อาจเป็น null |
| `mobileNumber` | `string` | อาจไม่ส่งเมื่อ zero value |
| `email` | `string` | อาจไม่ส่งเมื่อ zero value |
| `certEndDate` | `*string` | ส่ง field; pointer อาจเป็น null |
| `address` | array of [renewal.ValidateAddress](SCHEMAS.md#renewalvalidateaddress) | ส่ง field; pointer อาจเป็น null |
| `items` | array of [renewal.ValidateItem](SCHEMAS.md#renewalvalidateitem) | ส่ง field; pointer อาจเป็น null |

## requestitem.File

Source: [internal/domain/requestitem/model.go](../../internal/domain/requestitem/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `fileId` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentType` | `string` | ส่ง field; pointer อาจเป็น null |
| `slotCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `filePath` | `string` | อาจไม่ส่งเมื่อ zero value |
| `originalFileName` | `string` | ส่ง field; pointer อาจเป็น null |
| `mimeType` | `string` | ส่ง field; pointer อาจเป็น null |
| `fileSize` | `int64` | ส่ง field; pointer อาจเป็น null |
| `fileUploadedAt` | `string` | อาจไม่ส่งเมื่อ zero value |
| `checkResult` | `string` | อาจไม่ส่งเมื่อ zero value |
| `checkNote` | `string` | อาจไม่ส่งเมื่อ zero value |
| `isUpdated` | `bool` | อาจไม่ส่งเมื่อ zero value |

## requestitem.Item

Source: [internal/domain/requestitem/model.go](../../internal/domain/requestitem/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `id` | `string` | ส่ง field; pointer อาจเป็น null |
| `profileRequestItemId` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentCode` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentMasterRequestItemCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `storageScope` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentType` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentName` | `string` | อาจไม่ส่งเมื่อ zero value |
| `documentStatus` | `string` | ส่ง field; pointer อาจเป็น null |
| `sortOrder` | `int` | ส่ง field; pointer อาจเป็น null |
| `fileUploaded` | `int` | ส่ง field; pointer อาจเป็น null |
| `filePath` | `string` | อาจไม่ส่งเมื่อ zero value |
| `fileUploadedAt` | `string` | อาจไม่ส่งเมื่อ zero value |
| `checkResult` | `string` | อาจไม่ส่งเมื่อ zero value |
| `checkNote` | `string` | อาจไม่ส่งเมื่อ zero value |
| `files` | array of [requestitem.File](SCHEMAS.md#requestitemfile) | อาจไม่ส่งเมื่อ zero value |

## requestitem.PreviewResponse

Source: [internal/domain/requestitem/model.go](../../internal/domain/requestitem/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `documentMasterRequestItemCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentType` | `string` | อาจไม่ส่งเมื่อ zero value |
| `files` | array of [requestitem.File](SCHEMAS.md#requestitemfile) | ส่ง field; pointer อาจเป็น null |

## requestitem.UploadResponse

Source: [internal/domain/requestitem/model.go](../../internal/domain/requestitem/model.go)

| JSON field | Type | Serialization |
|---|---|---|
| `requestId` | `string` | อาจไม่ส่งเมื่อ zero value |
| `requestNo` | `string` | อาจไม่ส่งเมื่อ zero value |
| `profileRequestItemId` | `string` | อาจไม่ส่งเมื่อ zero value |
| `requestItemFileId` | `string` | อาจไม่ส่งเมื่อ zero value |
| `itemCode` | `string` | ส่ง field; pointer อาจเป็น null |
| `storageScope` | `string` | ส่ง field; pointer อาจเป็น null |
| `documentType` | `string` | ส่ง field; pointer อาจเป็น null |
| `complete` | `bool` | ส่ง field; pointer อาจเป็น null |
| `files` | array of [requestitem.File](SCHEMAS.md#requestitemfile) | ส่ง field; pointer อาจเป็น null |

