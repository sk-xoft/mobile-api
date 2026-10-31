-- Destructive cleanup for document renewal request data.
-- Intended for local/dev/test reset only. Do not run on shared or production DB
-- without explicit approval and backup.
--
-- The requested core tables are:
--   m_document_transaction
--   m_document_request_items
--   m_document_request
--
-- Dependent child tables are deleted first so foreign key checks can stay enabled.

START TRANSACTION;

DELETE FROM m_delivery;
DELETE FROM m_dept_submission;
DELETE FROM m_document_request_delivery_address;
DELETE FROM m_payment_transaction;
DELETE FROM m_document_transaction;
DELETE FROM m_document_request_item_files;
DELETE FROM m_document_request_items;
DELETE FROM m_document_request;

COMMIT;
