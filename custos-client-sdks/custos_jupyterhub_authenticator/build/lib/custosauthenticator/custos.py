#  Licensed to the Apache Software Foundation (ASF) under one or more
#  contributor license agreements.  See the NOTICE file distributed with
#  this work for additional information regarding copyright ownership.
#  The ASF licenses this file to You under the Apache License, Version 2.0
#  (the "License"); you may not use this file except in compliance with
#  the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#


"""
Custos Authenticator to use  OAuth2 with JupyterHub
"""
import base64
import os
from urllib.parse import urlencode
import logging

import os
from jupyterhub.auth import LocalAuthenticator
from tornado import web
from tornado.httpclient import HTTPRequest
from tornado.web import HTTPError
from tornado.httputil import url_concat
from traitlets import Bool
from traitlets import default
from traitlets import List
from traitlets import Unicode
from traitlets import validate
from tornado.log import app_log

from oauthenticator.oauth2 import OAuthenticator
from oauthenticator.oauth2 import OAuthLoginHandler


class CustosLoginHandler(OAuthLoginHandler):
    """See //https://airavata.apache.org/custos/ for general information."""

    def authorize_redirect(self, *args, **kwargs):
        """Add idp, skin to redirect params"""
        extra_params = kwargs.setdefault('extra_params', {})
        extra_params["kc_idp_hint"] = 'oidc'
        return super().authorize_redirect(*args, **kwargs)


class CustosOAuthenticator(OAuthenticator):
    custos_host = Unicode(os.environ.get("CUSTOS_HOST") or "prod.custos.usecustos.org", config=True)

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.login_service = "Custos Login"
        self.login_handler = CustosLoginHandler
        iam_host = ''
        if self.custos_host == 'dev.custos.usecustos.org':
            iam_host = "dev.keycloak.usecustos.org"
        elif self.custos_host == 'staging.custos.usecustos.org':
            iam_host = "staging.keycloak.usecustos.org"
        elif self.custos_host == 'prod.custos.usecustos.org':
            iam_host = "prod.keycloak.usecustos.org"
        x = super().client_id.split("-")
        tenant_id = x[len(x) - 1]
        self.iam_uri = "https://{}/auth/realms/{}/protocol/openid-connect/".format(iam_host, tenant_id)
        self.group_uri = "https://{}/apiserver/group-management/v1.0.0/user/group/memberships".format(self.custos_host)

    @default("authorize_url")
    def _authorize_url_default(self):
        return "{}auth".format(self.iam_uri)

    @default("token_url")
    def _token_url_default(self):
        return "https://{}/apiserver/identity-management/v1.0.0/token".format(self.custos_host)

    scope = List(
        Unicode(),
        default_value=['openid', 'email', 'org.cilogon.userinfo'],
        config=True,
        help="""The OAuth scopes to request.
        See cilogon_scope.md for details.
        At least 'openid' is required.
        """, )

    allowed_groups = List(
        config=True, )

    @validate('scope')
    def _validate_scope(self, proposal):
        """ensure openid is requested"""

        if 'openid' not in proposal.value:
            return ['openid'] + proposal.value
        return proposal.value

    async def authenticate(self, handler, data=None):
        """We set up auth_state based on additional Custos info if we
            receive it.
            """

        code = handler.get_argument("code")

        authS = "{}:{}".format(self.client_id, self.client_secret)
        tokenByte = authS.encode('utf-8')
        encodedBytes = base64.b64encode(tokenByte)
        auth_string = encodedBytes.decode('utf-8')
        headers = {"Accept": "application/json", "User-Agent": "JupyterHub",
                   "Authorization": "Bearer {}".format(auth_string)}

        params = dict(
            client_id=self.client_id,
            client_secret=self.client_secret,
            redirect_uri=self.oauth_callback_url,
            code=code,
            grant_type='authorization_code',
        )

        url = url_concat(self.token_url, params)

        req = HTTPRequest(url, headers=headers, method="POST", body='')

        token_response = await self.fetch(req)
        access_token = token_response['access_token']

        # Determine who the logged in user is
        params = dict(access_token=access_token)
        req = HTTPRequest(
            url_concat("https://{}/apiserver/identity-management/v1.0.0/user".format(self.custos_host), params),
            headers=headers,
        )
        resp_json = await  self.fetch(req)

        userdict = {"name": resp_json['username']}
        # Now we set up auth_state
        userdict["auth_state"] = auth_state = {}
        # Save the token response and full Custos reply in auth state
        # These can be used for user provisioning
        #  in the Lab/Notebook environment.
        auth_state['token_response'] = token_response
        # store the whole user model in auth_state.custos_user
        # keep access_token as well, in case anyone was relying on it
        auth_state['access_token'] = access_token
        auth_state['custos_user'] = resp_json
        return userdict

    async def pre_spawn_start(self, user, spawner):
        """Pass upstream_token to spawner via environment variable"""
        app_log.debug("Calling pre_spawn_start")
        auth_state = await user.get_auth_state()
        if not auth_state:
            # auth_state not enabled
            app_log.debug("Auth state not enabled")
            return

        authentication_status = await self.is_user_authorized_to_spawn_server(user.name)
        if not authentication_status:
            msg = "User {} is not authorized to start a server".format(user.name)
            raise HTTPError(401, msg)
        spawner.environment['UPSTREAM_TOKEN'] = auth_state['access_token']

    async def is_user_authorized_to_spawn_server(self, username):

        authS = "{}:{}".format(self.client_id, self.client_secret)
        tokenByte = authS.encode('utf-8')
        encodedBytes = base64.b64encode(tokenByte)
        auth_string = encodedBytes.decode('utf-8')
        headers = {"Accept": "application/json", "User-Agent": "JupyterHub",
                   "Authorization": "Bearer {}".format(auth_string)}

        # Determine who the logged in user is
        key = ['profile.username']
        value = [username]

        params = dict(zip(key, value))

        url = url_concat(self.group_uri, params)

        req = HTTPRequest(url, headers=headers, method="GET")

        group_response = await self.fetch(req)

        user_groups = group_response['groups']
        matched = False

        for group in user_groups:
            if group['id']  in self.allowed_groups:
                matched = True
                return matched

        return matched


