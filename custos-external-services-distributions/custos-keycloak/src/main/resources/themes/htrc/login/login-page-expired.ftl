<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
    <#elseif section = "form">
        <style>
            #kc-instruction {
                position: relative;
                z-index: 1;
                background: #ffffff;
                max-width: 460px;
                margin: 0 auto 10px;
                padding: 45px;
                border-radius: 2px;
                box-shadow: 0 0 20px 0 rgba(0, 0, 0, 0.2), 0 5px 5px 0 rgba(0, 0, 0, 0.24);
                color: red;
            }

            #kc-instruction-header > p {
                font-weight: bold;
                font-size: 1.4rem;
                border-bottom: #dddddd 1px solid;
                padding-bottom: 5px;
            }
        </style>
        <div id="kc-instruction">
            <div id="kc-instruction-header">
                <p>
                    Page has expired
                </p>
            </div>
            <div id="kc-instruction-message">
                <p><a id="backToApplication"
                      href="${client.baseUrl}">${kcSanitize(msg("backToApplication"))?no_esc}</a></p>
<#--                <ul>-->
<#--                    <li>To restart the login process <a id="loginRestartLink" href="${url.loginRestartFlowUrl}">Click here</a> .</li>-->
<#--                    <li>To continue the login process <a id="loginContinueLink" href="${url.loginAction}">Click here</a> .</li>-->
<#--                </ul>-->
            </div>
        </div>

    </#if>
</@layout.registrationLayout>
