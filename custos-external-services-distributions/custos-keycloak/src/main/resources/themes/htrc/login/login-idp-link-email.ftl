<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
<#--        ${msg("emailLinkIdpTitle", idpAlias)}-->
    <#elseif section = "form">
        <div id="kc-link-confirm">
            <div id="kc-link-confirm-header">
                <p>Link HTRC Analytics Account</p>
            </div>

            <p id="instruction1" class="htrc-body-text">
                An email has been sent to the email address associated with your HTRC Analytics account.<br><br>
                Follow the instructions in that email to link your institution's new authentication credentials to your HTRC account.
                <#--            ${msg("emailLinkIdp1", idpAlias, brokerContext.username, realm.displayName)}-->
            </p>

            <p id="instruction2" class="htrc-body-text">
                Haven't received a verification email? Click <a href="${url.loginAction}">here</a> to resend that email.
            </p>
        </div>
<#--        <p id="instruction3" class="instruction">-->
<#--            ${msg("emailLinkIdp4")} <a href="${url.loginAction}">${msg("doClickHere")}</a> ${msg("emailLinkIdp5")}-->
<#--        </p>-->
    </#if>
</@layout.registrationLayout>