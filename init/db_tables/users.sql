SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int(8) NOT NULL AUTO_INCREMENT,
  `uuid` varchar(37) UNIQUE NOT NULL,
  `login` varchar(127) UNIQUE NOT NULL,
  `password` varchar(127) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `users` (`id`, `uuid`, `login`, `password`) VALUES
(1,	'ffffffff-ffff-ffff-ffff-ffffffffffff',	'admin',	'rootroot'),
(2,	'12345678-9abc-def1-2345-6789abcdef12',	'test_user',	'useruser');
