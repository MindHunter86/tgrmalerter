-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

ALTER TABLE `tgrmAlerter`.`users` 
DROP COLUMN `chat_id`,
CHANGE COLUMN `registered` `registered` TINYINT(1) NOT NULL ,
ADD COLUMN `used_id` INT(10) UNSIGNED NOT NULL AFTER `phone`,
ADD COLUMN `username` VARCHAR(32) NOT NULL AFTER `used_id`,
ADD COLUMN `first_name` VARCHAR(24) NOT NULL AFTER `registered`,
ADD COLUMN `last_name` VARCHAR(32) NOT NULL AFTER `first_name`,
ADD UNIQUE INDEX `userid_UNIQUE` (`username` ASC),
ADD UNIQUE INDEX `used_id_UNIQUE` (`used_id` ASC),
DROP INDEX `chat_id_UNIQUE` ;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;

