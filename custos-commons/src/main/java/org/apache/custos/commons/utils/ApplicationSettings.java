package org.apache.custos.commons.utils;

import org.apache.custos.commons.exceptions.ApplicationSettingsException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.*;
import java.util.regex.Pattern;

public class ApplicationSettings {
    protected static ApplicationSettings INSTANCE;
    private Exception propertyLoadException;
    public static final String SERVER_PROPERTIES="custos-server.properties";
    public static final String CUSTOS_CONFIG_DIR = "custos.config.dir";
    public static String ADDITIONAL_SETTINGS_FILES = "external.settings";
    protected static final String TRUST_STORE_PATH="trust.store";
    protected static final String TRUST_STORE_PASSWORD="trust.store.password";
    private final static Logger logger = LoggerFactory.getLogger(ApplicationSettings.class);
    protected Properties properties = new Properties();
    // Profile Service Constants
    public static final String PROFILE_SERVICE_SERVER_HOST = "profile.service.server.host";
    public static final String PROFILE_SERVICE_SERVER_PORT = "profile.service.server.port";

    // Iam Server Constants
    public static final String IAM_SERVER_URL = "iam.server.url";
    public static final String IAM_SERVER_SUPER_ADMIN_USERNAME = "iam.server.super.admin.username";
    public static final String IAM_SERVER_SUPER_ADMIN_PASSWORD = "iam.server.super.admin.password";
    {
        loadProperties();
    }
    private void loadProperties() {
        URL url = getPropertyFileURL();
        try {
            properties.load(url.openStream());
            logger.info("Settings loaded from "+url.toString());
            URL[] externalSettingsFileURLs = getExternalSettingsFileURLs();
            for (URL externalSettings : externalSettingsFileURLs) {
                mergeSettingsImpl(externalSettings.openStream());
                logger.info("External settings merged from "+url.toString());
            }
        } catch (Exception e) {
            propertyLoadException=e;
        }
    }
    public void mergeSettingsImpl(InputStream stream) throws IOException {
        Properties tmpProp = new Properties();
        tmpProp.load(stream);
        properties.putAll(tmpProp);
    }
    private URL[] getExternalSettingsFileURLs(){
        try {
            List<URL> externalSettingsFileURLs=new ArrayList<URL>();
            String externalSettingsFileNames = getSettingImpl(ADDITIONAL_SETTINGS_FILES);
            String[] externalSettingFiles = externalSettingsFileNames.split(",");
            for (String externalSettingFile : externalSettingFiles) {
                URL externalSettingFileURL = ApplicationSettings.loadFile(externalSettingFile);
                if (externalSettingFileURL==null){
                    logger.warn("Could not file external settings file "+externalSettingFile);
                }else{
                    externalSettingsFileURLs.add(externalSettingFileURL);
                }
            }
            return externalSettingsFileURLs.toArray(new URL[]{});
        } catch (ApplicationSettingsException e) {
            return new URL[]{};
        }
    }
    private URL getPropertyFileURL() {
        return ApplicationSettings.loadFile(SERVER_PROPERTIES);
    }
    private static URL loadFile(String fileName) {

        if(System.getProperty(CUSTOS_CONFIG_DIR) != null) {
            String custosConfigDir = System.getProperty(CUSTOS_CONFIG_DIR);
            try {
                custosConfigDir = custosConfigDir.endsWith(File.separator) ? custosConfigDir : custosConfigDir + File.separator;
                String filePath = custosConfigDir + fileName;

                File asfile  = new File(filePath);
                if (asfile.exists()) {

                    return asfile.toURI().toURL();
                }
            } catch (MalformedURLException e) {
                logger.error("Error parsing the file from custos.config.dir", custosConfigDir);
            }
        }

        return ApplicationSettings.class.getClassLoader().getResource(fileName);

    }
    protected static ApplicationSettings getInstance(){
        if (INSTANCE==null){
            INSTANCE=new ApplicationSettings();
        }
        return INSTANCE;
    }
    public static String getSetting(String key) throws ApplicationSettingsException {
        return getInstance().getSettingImpl(key);
    }
    public static String getSetting(String key, String defaultValue) {
        return getInstance().getSettingImpl(key,defaultValue);

    }
    public static void setSetting(String key, String value) throws ApplicationSettingsException{
        getInstance().properties.setProperty(key, value);
        getInstance().saveProperties();
    }
    private void saveProperties() throws ApplicationSettingsException{
        URL url = getPropertyFileURL();
        if (url.getProtocol().equalsIgnoreCase("file")){
            try {
                properties.store(new FileOutputStream(url.getPath()), Calendar.getInstance().toString());
            } catch (Exception e) {
                throw new ApplicationSettingsException(url.getPath(), e);
            }
        }else{
            logger.warn("Properties cannot be updated to location "+url.toString());
        }
    }
    private String getSettingImpl(String key, String defaultValue){
        try {
            return getSettingImpl(key);
        } catch (ApplicationSettingsException e) {
            //we'll ignore this error since a default value is provided
        }
        return defaultValue;
    }
    private String getSettingImpl(String key) throws ApplicationSettingsException{
        String rawValue=null;
        if (System.getProperties().containsKey(key)){
            rawValue=System.getProperties().getProperty(key);
        }else{
            validateSuccessfulPropertyFileLoad();
            if (properties.containsKey(key)){
                rawValue=properties.getProperty(key);
            }else{
                throw new ApplicationSettingsException(key);
            }
        }
        return deriveAbsoluteValueImpl(rawValue);
    }
    private String deriveAbsoluteValueImpl(String property){
        if (property!=null){
            Map<Integer, String> containedParameters = StringUtil.getContainedParameters(property);
            List<String> parametersAlreadyProcessed=new ArrayList<String>();
            for (String parameter : containedParameters.values()) {
                if (!parametersAlreadyProcessed.contains(parameter)) {
                    String parameterName = parameter.substring(2,parameter.length() - 1);
                    String parameterValue = getSetting(parameterName,parameter);
                    property = property.replaceAll(Pattern.quote(parameter), parameterValue);
                    parametersAlreadyProcessed.add(parameter);
                }
            }
        }
        return property;
    }
    private void validateSuccessfulPropertyFileLoad() throws ApplicationSettingsException{
        if (propertyLoadException!=null){
            throw new ApplicationSettingsException(propertyLoadException.getMessage(), propertyLoadException);
        }
    }
    public static String getTrustStorePath() throws ApplicationSettingsException {
        return getSetting(TRUST_STORE_PATH);
    }
    public static String getTrustStorePassword() throws ApplicationSettingsException {
        return getSetting(TRUST_STORE_PASSWORD);
    }
    public static String getIamServerUrl() throws ApplicationSettingsException {
        return getSetting(ServerSettings.IAM_SERVER_URL);
    }
    public static String getProfileServiceServerHost() throws ApplicationSettingsException {
        return getSetting(ServerSettings.PROFILE_SERVICE_SERVER_HOST);
    }

    public static String getProfileServiceServerPort() throws ApplicationSettingsException {
        return getSetting(ServerSettings.PROFILE_SERVICE_SERVER_PORT);
    }
}
