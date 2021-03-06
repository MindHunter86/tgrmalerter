-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

ALTER TABLE `tgrmAlerter`.`dispatch_reports` 
DROP COLUMN `dispatch_status`,
ADD COLUMN `status_error` VARCHAR(36) NULL DEFAULT NULL AFTER `status_code`,
ADD INDEX `fk_dispatch_error_idx` (`status_error` ASC);

ALTER TABLE `tgrmAlerter`.`dispatch_reports` 
ADD CONSTRAINT `fk_dispatch_error`
  FOREIGN KEY (`status_error`)
  REFERENCES `tgrmAlerter`.`errors` (`id`)
  ON DELETE RESTRICT
  ON UPDATE RESTRICT;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;

