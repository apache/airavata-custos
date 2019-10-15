--
-- Current Database: `profile_service`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `profile_service` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */;

USE `profile_service`;

--
-- Table structure for table `CONFIGURATION`
--

DROP TABLE IF EXISTS `CONFIGURATION`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `CONFIGURATION` (
  `CONFIG_KEY` varchar(255) NOT NULL,
  `CONFIG_VAL` varchar(255) NOT NULL,
  PRIMARY KEY (`CONFIG_KEY`,`CONFIG_VAL`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `CONFIGURATION`
--

LOCK TABLES `CONFIGURATION` WRITE;
/*!40000 ALTER TABLE `CONFIGURATION` DISABLE KEYS */;
INSERT INTO `CONFIGURATION` VALUES ('user_profile_catalog_version','0.17');
/*!40000 ALTER TABLE `CONFIGURATION` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_PROFILE`
--

DROP TABLE IF EXISTS `USER_PROFILE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_PROFILE` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `USER_ID` varchar(255) NOT NULL,
  `GATEWAY_ID` varchar(255) NOT NULL,
  `USER_MODEL_VERSION` varchar(255) DEFAULT NULL,
  `FIRST_NAME` varchar(255) DEFAULT NULL,
  `LAST_NAME` varchar(255) DEFAULT NULL,
  `MIDDLE_NAME` varchar(255) DEFAULT NULL,
  `NAME_PREFIX` varchar(255) DEFAULT NULL,
  `NAME_SUFFIX` varchar(255) DEFAULT NULL,
  `ORCID_ID` varchar(255) DEFAULT NULL,
  `COUNTRY` varchar(255) DEFAULT NULL,
  `HOME_ORGANIZATION` varchar(255) DEFAULT NULL,
  `ORIGINATION_AFFILIATION` varchar(255) DEFAULT NULL,
  `CREATION_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `LAST_ACCESS_TIME` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `VALID_UNTIL` datetime DEFAULT NULL,
  `STATE` varchar(255) DEFAULT NULL,
  `COMMENTS` text,
  `GPG_KEY` text,
  `TIME_ZONE` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_PROFILE`
--

LOCK TABLES `USER_PROFILE` WRITE;
/*!40000 ALTER TABLE `USER_PROFILE` DISABLE KEYS */;
INSERT INTO `USER_PROFILE` VALUES ('default-admin@default','default-admin','default','1.0','dim','Upe',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2019-02-25 18:46:43','2019-02-25 18:46:43','1969-12-31 19:00:00','ACTIVE',NULL,NULL,NULL);
/*!40000 ALTER TABLE `USER_PROFILE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `GATEWAY`
--

DROP TABLE IF EXISTS `GATEWAY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `GATEWAY` (
     `CUSTOS_INTERNAL_GATEWAY_ID` varchar(255) NOT NULL,
     `DECLINED_REASON` varchar(255) DEFAULT NULL,
     `GATEWAY_DOMAIN` varchar(255) DEFAULT NULL,
     `EMAIL_ADDRESS` varchar(255) DEFAULT NULL,
     `GATEWAY_ACRONYM` varchar(255) DEFAULT NULL,
     `GATEWAY_ADMIN_EMAIL` varchar(255) DEFAULT NULL,
     `GATEWAY_ADMIN_FIRST_NAME` varchar(255) DEFAULT NULL,
     `GATEWAY_ADMIN_LAST_NAME` varchar(255) DEFAULT NULL,
     `GATEWAY_APPROVAL_STATUS` varchar(255) DEFAULT NULL,
     `GATEWAY_ID` varchar(255) DEFAULT NULL,
     `GATEWAY_NAME` varchar(255) DEFAULT NULL,
     `GATEWAY_PUBLIC_ABSTRACT` varchar(255) DEFAULT NULL,
     `GATEWAY_URL` varchar(255) DEFAULT NULL,
     `IDENTITY_SERVER_PASSWORD_TOKEN` varchar(255) DEFAULT NULL,
     `IDENTITY_SERVER_USERNAME` varchar(255) DEFAULT NULL,
     `OAUTH_CLIENT_ID` varchar(255) DEFAULT NULL,
     `OAUTH_CLIENT_SECRET` varchar(255) DEFAULT NULL,
     `REQUEST_CREATION_TIME` bigint(20) DEFAULT NULL,
     `REQUESTER_USERNAME` varchar(255) DEFAULT NULL,
     `GATEWAY_REVIEW_PROPOSAL_DESCRIPTION` varchar(255) DEFAULT NULL,
     PRIMARY KEY (`CUSTOS_INTERNAL_GATEWAY_ID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `GATEWAY`
--

LOCK TABLES `GATEWAY` WRITE;
/*!40000 ALTER TABLE `GATEWAY` DISABLE KEYS */;
INSERT INTO `GATEWAY` VALUES ('default',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'APPROVED','default','default',NULL,NULL,NULL,NULL,'pga','9790c8c4-7d9b-4ccc-a820-ca5aac38d2ad','2019-02-25 18:40:06',NULL, NULL);
/*!40000 ALTER TABLE `GATEWAY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `NSF_DEMOGRAPHIC`
--

DROP TABLE IF EXISTS `NSF_DEMOGRAPHIC`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `NSF_DEMOGRAPHIC` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `GENDER` varchar(255) NOT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`),
  CONSTRAINT `nsf_demographic_ibfk_1` FOREIGN KEY (`CUSTOS_INTERNAL_USER_ID`) REFERENCES `USER_PROFILE` (`CUSTOS_INTERNAL_USER_ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `NSF_DEMOGRAPHIC`
--

LOCK TABLES `NSF_DEMOGRAPHIC` WRITE;
/*!40000 ALTER TABLE `NSF_DEMOGRAPHIC` DISABLE KEYS */;
/*!40000 ALTER TABLE `NSF_DEMOGRAPHIC` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `NSF_DEMOGRAPHIC_DISABILITY`
--

DROP TABLE IF EXISTS `NSF_DEMOGRAPHIC_DISABILITY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `NSF_DEMOGRAPHIC_DISABILITY` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `DISABILITY` varchar(255) NOT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`,`DISABILITY`),
  CONSTRAINT `nsf_demographic_disability_ibfk_1` FOREIGN KEY (`CUSTOS_INTERNAL_USER_ID`) REFERENCES `NSF_DEMOGRAPHIC` (`CUSTOS_INTERNAL_USER_ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `NSF_DEMOGRAPHIC_DISABILITY`
--

LOCK TABLES `NSF_DEMOGRAPHIC_DISABILITY` WRITE;
/*!40000 ALTER TABLE `NSF_DEMOGRAPHIC_DISABILITY` DISABLE KEYS */;
/*!40000 ALTER TABLE `NSF_DEMOGRAPHIC_DISABILITY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `NSF_DEMOGRAPHIC_ETHNICITY`
--

DROP TABLE IF EXISTS `NSF_DEMOGRAPHIC_ETHNICITY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `NSF_DEMOGRAPHIC_ETHNICITY` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `ETHNICITY` varchar(255) NOT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`,`ETHNICITY`),
  CONSTRAINT `nsf_demographic_ethnicity_ibfk_1` FOREIGN KEY (`CUSTOS_INTERNAL_USER_ID`) REFERENCES `NSF_DEMOGRAPHIC` (`CUSTOS_INTERNAL_USER_ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `NSF_DEMOGRAPHIC_ETHNICITY`
--

LOCK TABLES `NSF_DEMOGRAPHIC_ETHNICITY` WRITE;
/*!40000 ALTER TABLE `NSF_DEMOGRAPHIC_ETHNICITY` DISABLE KEYS */;
/*!40000 ALTER TABLE `NSF_DEMOGRAPHIC_ETHNICITY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `NSF_DEMOGRAPHIC_RACE`
--

DROP TABLE IF EXISTS `NSF_DEMOGRAPHIC_RACE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `NSF_DEMOGRAPHIC_RACE` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `RACE` varchar(255) NOT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`,`RACE`),
  CONSTRAINT `nsf_demographic_race_ibfk_1` FOREIGN KEY (`CUSTOS_INTERNAL_USER_ID`) REFERENCES `NSF_DEMOGRAPHIC` (`CUSTOS_INTERNAL_USER_ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `NSF_DEMOGRAPHIC_RACE`
--

LOCK TABLES `NSF_DEMOGRAPHIC_RACE` WRITE;
/*!40000 ALTER TABLE `NSF_DEMOGRAPHIC_RACE` DISABLE KEYS */;
/*!40000 ALTER TABLE `NSF_DEMOGRAPHIC_RACE` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_PROFILE_EMAIL`
--

DROP TABLE IF EXISTS `USER_PROFILE_EMAIL`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_PROFILE_EMAIL` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `EMAIL` varchar(255) NOT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`,`EMAIL`),
  CONSTRAINT `user_profile_email_ibfk_1` FOREIGN KEY (`CUSTOS_INTERNAL_USER_ID`) REFERENCES `USER_PROFILE` (`CUSTOS_INTERNAL_USER_ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_PROFILE_EMAIL`
--

LOCK TABLES `USER_PROFILE_EMAIL` WRITE;
/*!40000 ALTER TABLE `USER_PROFILE_EMAIL` DISABLE KEYS */;
INSERT INTO `USER_PROFILE_EMAIL` VALUES ('default-admin@default','dimuthu.upeksha2@gmail.com');
/*!40000 ALTER TABLE `USER_PROFILE_EMAIL` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_PROFILE_LABELED_URI`
--

DROP TABLE IF EXISTS `USER_PROFILE_LABELED_URI`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_PROFILE_LABELED_URI` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `LABELED_URI` varchar(255) NOT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`,`LABELED_URI`),
  CONSTRAINT `user_profile_labeled_uri_ibfk_1` FOREIGN KEY (`CUSTOS_INTERNAL_USER_ID`) REFERENCES `USER_PROFILE` (`CUSTOS_INTERNAL_USER_ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_PROFILE_LABELED_URI`
--

LOCK TABLES `USER_PROFILE_LABELED_URI` WRITE;
/*!40000 ALTER TABLE `USER_PROFILE_LABELED_URI` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_PROFILE_LABELED_URI` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_PROFILE_NATIONALITY`
--

DROP TABLE IF EXISTS `USER_PROFILE_NATIONALITY`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_PROFILE_NATIONALITY` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `NATIONALITY` varchar(255) NOT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`,`NATIONALITY`),
  CONSTRAINT `user_profile_nationality_ibfk_1` FOREIGN KEY (`CUSTOS_INTERNAL_USER_ID`) REFERENCES `USER_PROFILE` (`CUSTOS_INTERNAL_USER_ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_PROFILE_NATIONALITY`
--

LOCK TABLES `USER_PROFILE_NATIONALITY` WRITE;
/*!40000 ALTER TABLE `USER_PROFILE_NATIONALITY` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_PROFILE_NATIONALITY` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `USER_PROFILE_PHONE`
--

DROP TABLE IF EXISTS `USER_PROFILE_PHONE`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `USER_PROFILE_PHONE` (
  `CUSTOS_INTERNAL_USER_ID` varchar(255) NOT NULL,
  `PHONE` varchar(255) NOT NULL,
  PRIMARY KEY (`CUSTOS_INTERNAL_USER_ID`,`PHONE`),
  CONSTRAINT `user_profile_phone_ibfk_1` FOREIGN KEY (`CUSTOS_INTERNAL_USER_ID`) REFERENCES `USER_PROFILE` (`CUSTOS_INTERNAL_USER_ID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `USER_PROFILE_PHONE`
--

LOCK TABLES `USER_PROFILE_PHONE` WRITE;
/*!40000 ALTER TABLE `USER_PROFILE_PHONE` DISABLE KEYS */;
/*!40000 ALTER TABLE `USER_PROFILE_PHONE` ENABLE KEYS */;
UNLOCK TABLES;