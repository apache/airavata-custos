<#import "template.ftl" as layout>
<@layout.registrationLayout displayInfo=social.displayInfo; section>
    <#if section = "title">
        Welcome to HTRC!
    <#elseif section = "header">
        <link href="https://fonts.googleapis.com/css?family=Roboto" rel="stylesheet">
        <link href="${url.resourcesPath}/img/favicon.png" rel="icon"/>
        <script>
            function togglePassword() {
                var x = document.getElementById("password");
                var v = document.getElementById("vi");
                if (x.type === "password") {
                    x.type = "text";
                    v.src = "${url.resourcesPath}/img/eye.png";
                } else {
                    x.type = "password";
                    v.src = "${url.resourcesPath}/img/eye-off.png";
                }
            }
        </script>
        <div class="welcome">
            <div class="welcome-text">
                <p>Welcome!</p>
            </div>
        </div>
    <#elseif section = "form">
        <div class="login">
        <#if realm.password>
            <div class="login-form">
                <h3>Sign in to <span class="htrc-text">HathiTrust Research Center</span></h3>
                <#if message?has_content>
                    <div class="alert alert-${message.type}">
                        <#if message.type = 'success'><span class="${properties.kcFeedbackSuccessIcon!}"></span></#if>
                        <#if message.type = 'warning'><span class="${properties.kcFeedbackWarningIcon!}"></span></#if>
                        <#if message.type = 'error'><span class="${properties.kcFeedbackErrorIcon!}"></span></#if>
                        <#if message.type = 'info'><span class="${properties.kcFeedbackInfoIcon!}"></span></#if>
                        <#if message.summary?contains("Account is disabled")>
                            <span class="message-text">Your account is locked.<br/>Please contact <a href="mailto:htrc-help@hathitrust.org">htrc-help@hathitrust.org</a>.</span>
                        <#else>
                            <span class="message-text">${message.summary?no_esc}</span>
                        </#if>
                    </div>
                </#if>
               <form id="kc-form-login" class="form" onsubmit="login.disabled = true; return true;" action="${url.loginAction}" method="post">
                <input id="username" class="login-field" placeholder="User name" type="text" name="username" tabindex="1">
                <input id="password" class="login-field" placeholder="Password" type="password" name="password" tabindex="2">
                <button type="submit" tabindex="3">Sign In</button>
                </form>
                <div class="link-group">
                    <a id="passwordRecoverLink" href="${client.baseUrl}\passwordresetrequestpage">Forgot Password?</a> | <a id="usernameRecoverLink" href="${client.baseUrl}\useridrequestpage">Forgot Username?</a> | <a id="userSignUpLink" href="${client.baseUrl}\signuppage">Create Account</a>
                </div>
            </div>
        </#if>
        <#if social.providers??>
            <p class="para">${msg("selectAlternative")}</p>
            <div id="social-providers">
                <#list social.providers as p>
                <input class="social-link-style" type="button" onclick="location.href='${p.loginUrl}';" value="${p.displayName}"/>
                </#list>
            </div>
        </#if>
        <div class="login-footer">
            <p>HathiTrust Research Center | © <script>document.write(new Date().getFullYear());</script></p>
        </div>
        </div>
    </#if>
</@layout.registrationLayout>
