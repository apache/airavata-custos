<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">

    <#elseif section = "form">
        <div id="kc-link-confirm">
            <div id="kc-link-confirm-header">
                <p>${msg("emailVerifyTitle")}</p>
            </div>
            <p class="htrc-inst-text">
                ${msg("emailVerifyInstruction1")}
            </p>
            <p class="htrc-inst-text">
                ${msg("emailVerifyInstruction2")} <a href="${url.loginAction}">${msg("doClickHere")}</a> ${msg("emailVerifyInstruction3")}
            </p>
        </div>
    </#if>
</@layout.registrationLayout>