-- MySQL Workbench Synchronization

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL';

CREATE TABLE IF NOT EXISTS `tgrmAlerter`.`dispatch_reports` (
  `id` VARCHAR(36) NOT NULL,
  `request_id` VARCHAR(36) NOT NULL,
  `recipient` VARCHAR(10) NOT NULL,
  `message` VARCHAR(256) NOT NULL,
  `dispatch_status` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0,
  `status_code` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `id_UNIQUE` (`id` ASC),
  INDEX `fk_request_id_idx` (`request_id` ASC),
  INDEX `fk_recipient_idx` (`recipient` ASC),
  CONSTRAINT `fk_dispatch_records_request_id`
    FOREIGN KEY (`request_id`)
    REFERENCES `tgrmAlerter`.`requests` (`id`)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT,
  CONSTRAINT `fk_recipient`
    FOREIGN KEY (`recipient`)
    REFERENCES `tgrmAlerter`.`users` (`phone`)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;

