-- +migrate Up
ALTER TABLE `credentialusage` 
CHANGE COLUMN `date` `created` DATETIME NOT NULL ,
ADD COLUMN `updated` DATETIME NULL AFTER `created`,
ADD COLUMN `expired` TINYINT(1) NOT NULL DEFAULT 0 AFTER `updated`;

-- +migrate Down
ALTER TABLE `credentialusage` 
DROP COLUMN `expired`,
DROP COLUMN `updated`,
CHANGE COLUMN `created` `date` DATETIME NOT NULL ;
