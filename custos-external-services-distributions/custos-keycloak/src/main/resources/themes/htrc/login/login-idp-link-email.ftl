<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <style>
        #kc-link-confirm {
            position: relative;
            z-index: 1;
            background: #ffffff;
            max-width: 460px;
            margin: 0 auto 10px;
            padding: 45px;
            border-radius: 2px;
            box-shadow: 0 0 20px 0 rgba(0, 0, 0, 0.2), 0 5px 5px 0 rgba(0, 0, 0, 0.24);
            color: #ff6f01;
        }

        #kc-link-confirm-header > p {
            font-weight: bold;
            font-size: 1.4rem;
            border-bottom: #dddddd 1px solid;
            padding-bottom: 5px;
        }
    </style>
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