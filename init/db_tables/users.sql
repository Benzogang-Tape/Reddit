SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int(8) NOT NULL AUTO_INCREMENT,
  `uuid` varchar(37) NOT NULL
  `login` varchar(127) NOT NULL,
  `password` varchar(127) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `users` (`id`, `uuid`, `login`, `password`) VALUES
(1,	'ffffffff-ffff-ffff-ffff-ffffffffffff',	'admin',	'root'),
(2,	'8f4a45c5-e8b8-46d6-9c7f-2ba54447175d',	'test_user',	'user');
