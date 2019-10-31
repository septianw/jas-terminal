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

ALTER TABLE `credentialusage` 
ADD COLUMN `credentialusageid` INT NOT NULL AUTO_INCREMENT FIRST,
DROP PRIMARY KEY,
ADD PRIMARY KEY (`credentialusageid`, `clientcredential_clientid`, `terminal_terminalid`),
ADD UNIQUE INDEX `credentialusageid_UNIQUE` (`credentialusageid` ASC);

ALTER TABLE `accesstoken` 
ADD COLUMN `credentialusage_credentialusageid` INT NULL DEFAULT 0 AFTER `tokenusage_usageid`;

ALTER TABLE `refreshtoken` 
ADD COLUMN `credentialusage_credentialusageid` INT NULL DEFAULT 0 AFTER `tokenusage_usageid`;




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

ALTER TABLE `credentialusage` 
DROP COLUMN `credentialusageid`,
DROP PRIMARY KEY,
ADD PRIMARY KEY (`clientcredential_clientid`, `terminal_terminalid`),
DROP INDEX `credentialusageid_UNIQUE` ;

ALTER TABLE `accesstoken` 
DROP COLUMN `credentialusage_credentialusageid`;

ALTER TABLE `refreshtoken` 
DROP COLUMN `credentialusage_credentialusageid`;


