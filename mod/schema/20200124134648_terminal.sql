-- +migrate Up
ALTER TABLE `accesstoken` 
CHANGE COLUMN `timeout` `timeout` BIGINT(20) NULL DEFAULT NULL ;
ALTER TABLE `refreshtoken` 
CHANGE COLUMN `timeout` `timeout` BIGINT(20) NULL DEFAULT NULL ;


-- +migrate Down
ALTER TABLE `accesstoken` 
CHANGE COLUMN `timeout` `timeout` INT(11) NULL DEFAULT NULL ;
ALTER TABLE `refreshtoken` 
CHANGE COLUMN `timeout` `timeout` INT(11) NULL DEFAULT NULL ;