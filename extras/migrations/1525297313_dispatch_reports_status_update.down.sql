-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

ALTER TABLE `tgrmAlerter`.`dispatch_reports` 
DROP FOREIGN KEY `fk_dispatch_error`;

ALTER TABLE `tgrmAlerter`.`dispatch_reports` 
DROP COLUMN `status_error`,
ADD COLUMN `dispatch_status` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 AFTER `message`,
DROP INDEX `fk_dispatch_error_idx` ;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
