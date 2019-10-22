-- +migrate Up
CREATE TABLE `terminal` (
  `terminalid` varchar(48) COLLATE utf8_unicode_ci NOT NULL,
  `name` varchar(225) COLLATE utf8_unicode_ci DEFAULT NULL,
  `deleted` TINYINT(1) NOT NULL DEFAULT 0,
  `location_locid` int(11) NOT NULL,
  PRIMARY KEY (`terminalid`),
  CONSTRAINT unique_terminal_terminalid UNIQUE (terminalid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `clientcredential` (
  `clientid` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `clientsecret` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`clientid`),
  CONSTRAINT unique_clientcredential UNIQUE (clientid, clientsecret)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `credentialusage` (
  `clientcredential_clientid` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `terminal_terminalid` varchar(48) COLLATE utf8_unicode_ci NOT NULL,
  `date` datetime NOT NULL,
  PRIMARY KEY (`clientcredential_clientid`,`terminal_terminalid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `clientcredentialpermission` (
  `clientcredential_clientid` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `permission_permid` int(11) NOT NULL,
  PRIMARY KEY (`clientcredential_clientid`,`permission_permid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `tokenusage` (
  `usageid` int(11) NOT NULL AUTO_INCREMENT,
  `user_uid` int(11) NOT NULL,
  `user_uname` varchar(225) COLLATE utf8_unicode_ci NOT NULL,
  `terminal_terminalid` varchar(48) COLLATE utf8_unicode_ci NOT NULL,
  `date` datetime NOT NULL,
  PRIMARY KEY (`usageid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `accesstoken` (
  `token` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `timeout` int(11) DEFAULT NULL,
  `used` tinyint(1) NOT NULL DEFAULT 0,
  `tokenusage_usageid` int(11) DEFAULT NULL,
  PRIMARY KEY (`token`),
  CONSTRAINT unique_accesstoken_token UNIQUE (token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

CREATE TABLE `refreshtoken` (
  `token` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `timeout` int(11) DEFAULT NULL,
  `used` tinyint(1) NOT NULL DEFAULT 0,
  `tokenusage_usageid` int(11) DEFAULT NULL,
  PRIMARY KEY (`token`),
  CONSTRAINT unique_refreshtoken_token UNIQUE (token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS `terminal`;
DROP TABLE IF EXISTS `clientcredential`;
DROP TABLE IF EXISTS `credentialusage`;
DROP TABLE IF EXISTS `clientcredentialpermission`;
DROP TABLE IF EXISTS `tokenusage`;
DROP TABLE IF EXISTS `accesstoken`;
DROP TABLE IF EXISTS `refreshtoken`;