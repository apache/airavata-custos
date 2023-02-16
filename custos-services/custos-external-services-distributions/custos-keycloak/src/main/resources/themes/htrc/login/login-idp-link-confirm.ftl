<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
    <#elseif section = "form">
        <style>
            #kc-idp-link {
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

            #kc-idp-link-header > p {
                font-weight: bold;
                font-size: 1.4rem;
                border-bottom: #dddddd 1px solid;
                padding-bottom: 5px;
            }
        </style>
        <div id="kc-idp-link">
            <div id="kc-idp-link-header">
                <p>
                    Looks like you already have an account
                </p>
            </div>
            <div id="kc-idp-link-message">
                <p class="htrc-body-text">Your institution has made a change to their authentication credentials. Please click the button below to link the new credentials to your HTRC account.</p>
                <form id="kc-register-form" action="${url.loginAction}" method="post">
                <div class="${properties.kcFormGroupClass!}">
                    <#--                <button type="submit" class="${properties.kcButtonClass!} ${properties.kcButtonDefaultClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!}" name="submitAction" id="updateProfile" value="updateProfile">${msg("confirmLinkIdpReviewProfile")}</button>-->
                    <button type="submit" class="htrc-btn-confirm" name="submitAction" id="linkAccount" value="linkAccount">Update Account</button>
                </div>
                </form>
            </div>
        </div>
    </#if>
</@layout.registrationLayout>