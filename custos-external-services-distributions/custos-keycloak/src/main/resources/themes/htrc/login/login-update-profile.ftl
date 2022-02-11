<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
        <link href="https://fonts.googleapis.com/css?family=Roboto" rel="stylesheet">
        <div class="welcome">
            <div class="welcome-text">
                <p>Please Verify and Update Your Account Information</p>
            </div>
        </div>
    <#elseif section = "form">
        <div class="login update-info">
            <div class="login-form">
                <h3><span class="htrc-text">Account Information</span></h3>
        <form id="kc-update-profile-form" class="form" action="${url.loginAction}" method="post">
            <#if user.editUsernameAllowed>
                <div class="${properties.kcFormGroupClass!} ${messagesPerField.printIfExists('username',properties.kcFormGroupErrorClass!)}">
                    <div class="${properties.kcLabelWrapperClass!}">
                        <label for="username" class="${properties.kcLabelClass!}" hidden>${msg("username")}</label>
                    </div>
                    <div class="${properties.kcInputWrapperClass!}">
                        <input type="text" id="username" name="username" value="${(user.username!'')}" class="login-field" hidden/>
                    </div>
                </div>
            </#if>
            <div class="${properties.kcFormGroupClass!} ${messagesPerField.printIfExists('email',properties.kcFormGroupErrorClass!)}">
                <div class="${properties.kcLabelWrapperClass!}">
                    <label for="email" class="${properties.kcLabelClass!}">${msg("email")}</label>
                </div>
                <div class="${properties.kcInputWrapperClass!}">
                    <input type="text" id="email" name="email" value="${(user.email!'')}" class="login-field" placeholder="Please enter your institutional email"/>
                </div>
            </div>

            <div class="${properties.kcFormGroupClass!} ${messagesPerField.printIfExists('firstName',properties.kcFormGroupErrorClass!)}">
                <div class="${properties.kcLabelWrapperClass!}">
                    <label for="firstName" class="${properties.kcLabelClass!}">${msg("firstName")}</label>
                </div>
                <div class="${properties.kcInputWrapperClass!}">
                    <input type="text" id="firstName" name="firstName" value="${(user.firstName!'')}" class="login-field" />
                </div>
            </div>

            <div class="${properties.kcFormGroupClass!} ${messagesPerField.printIfExists('lastName',properties.kcFormGroupErrorClass!)}">
                <div class="${properties.kcLabelWrapperClass!}">
                    <label for="lastName" class="${properties.kcLabelClass!}">${msg("lastName")}</label>
                </div>
                <div class="${properties.kcInputWrapperClass!}">
                    <input type="text" id="lastName" name="lastName" value="${(user.lastName!'')}" class="login-field" />
                </div>
            </div>

            <div class="${properties.kcFormGroupClass!}">
                <div id="kc-form-options" class="${properties.kcFormOptionsClass!}">
                    <div class="${properties.kcFormOptionsWrapperClass!}">
                    </div>
                </div>

                <div id="kc-form-buttons" class="${properties.kcFormButtonsClass!}">
                    <input class="htrc-btn-confirm btn-update-info" type="submit" value="${msg("doSubmit")}" onclick="replaceUserNameValueFromEmail()" />
                </div>
            </div>
        </form>
            </div>
        </div>
    </#if>

    <script>
        function replaceUserNameValueFromEmail() {
            var userName = document.getElementById("username");
            var email = document.getElementById("email");
            userName.value = email.value;
        }
    </script>
</@layout.registrationLayout>
