<#import "template.ftl" as layout>
<@layout.registrationLayout displayInfo=true; section>
    <#if section = "header">
    <#elseif section = "form">
        <style>
            #kc-error {
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

            #kc-error-header > p {
                font-weight: bold;
                font-size: 1.4rem;
                border-bottom: #dddddd 1px solid;
                padding-bottom: 5px;
            }
        </style>
        <div id="kc-error">
            <div id="kc-error-header">
                <p>
                    We are sorry...
                </p>
            </div>
            <div id="kc-error-message">
                <p class="instruction">${message.summary}</p>
                <#if client?? && client.baseUrl?has_content>
                    <p><a id="backToApplication"
                          href="${client.baseUrl}">${kcSanitize(msg("backToApplication"))?no_esc}</a></p>
                </#if>
            </div>
        </div>
    </#if>
</@layout.registrationLayout>