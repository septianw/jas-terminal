-- +migrate Up
ALTER TABLE `clientcredential` 
ADD COLUMN `clientname` VARCHAR(255) NULL AFTER `clientsecret`,
ADD COLUMN `deleted` TINYINT(1) NULL DEFAULT 0 AFTER `clientname`,
ADD UNIQUE INDEX `clientname_UNIQUE` (`clientname` ASC);

ALTER TABLE `terminal` 
ADD UNIQUE INDEX `terminalname_UNIQUE` (`name` ASC);

ALTER TABLE `tokenusage` 
CHANGE COLUMN `date` `created` DATETIME NOT NULL ,
ADD COLUMN `updated` DATETIME NULL AFTER `created`,
ADD COLUMN `expired` TINYINT(1) NULL DEFAULT 0 AFTER `updated`;


-- +migrate Down
ALTER TABLE `clientcredential` 
DROP COLUMN `clientname`,
DROP COLUMN `deleted`,
DROP INDEX `clientname_UNIQUE` ;

ALTER TABLE `terminal` 
DROP INDEX `terminalname_UNIQUE` ;

ALTER TABLE `tokenusage` 
DROP COLUMN `expired`,
DROP COLUMN `updated`,
CHANGE COLUMN `created` `date` DATETIME NOT NULL ;

