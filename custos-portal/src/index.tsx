import { createRoot } from 'react-dom/client';
import App from './App';
import { extendTheme, ChakraProvider } from '@chakra-ui/react';
import { AuthProvider, AuthProviderProps } from 'react-oidc-context';
import { APP_REDIRECT_URI, BACKEND_URL, CLIENT_ID, TENANT_ID } from './lib/constants';
import { WebStorageStateStore } from 'oidc-client-ts';
import React, { useEffect, useState } from 'react';
import localOidcConfig from './lib/localOidcConfig.json';

const theme = extendTheme({
  colors: {
    default: {
      "default": "#1E1E1E",
      "secondary": "#757575",
      "tertiary": "#B3B3B3"
    },
    border: {
      neutral: {
        "default": "#303030",
        "secondary": "#767676",
        "tertiary": "#B2B2B2",
      }
    }
  },
});

const Index = () => {
  const [oidcConfig, setOidcConfig] = useState<AuthProviderProps | null>(null);

  useEffect(() => {
    const fetchOidcConfig = async () => {
      try {
        let data;
        if (!process.env.NODE_ENV || process.env.NODE_ENV === 'development') {
          data = localOidcConfig;
        } else {
          const response = await fetch(`${BACKEND_URL}/api/v1/identity-management/tenant/${TENANT_ID}/.well-known/openid-configuration`); // Replace with actual API endpoint
          data = await response.json();
        }

        // Determine redirect_uri based on environment
        const redirectUri = APP_REDIRECT_URI;

        // Create the OIDC config based on the fetched data
        const theConfig: AuthProviderProps = {
          authority: `${BACKEND_URL}/api/v1/identity-management/`,
          client_id: CLIENT_ID,
          redirect_uri: redirectUri,
          response_type: 'code',
          scope: 'openid email',
          metadata: {
            authorization_endpoint: data.authorization_endpoint,
            token_endpoint: data.token_endpoint,
            revocation_endpoint: data.revocation_endpoint,
            introspection_endpoint: data.introspection_endpoint,
            userinfo_endpoint: data.userinfo_endpoint,
            jwks_uri: data.jwks_uri,
          },
          userStore: new WebStorageStateStore({ store: window.localStorage }),
          automaticSilentRenew: true,
        };

        setOidcConfig(theConfig);
      } catch (error) {
        console.error('Error fetching OIDC config:', error);
      }
    };

    fetchOidcConfig();
  }, []);

  if (!oidcConfig) {
    return <div>Loading OIDC configuration...</div>; // Loading state while config is fetched
  }

  return (
    <ChakraProvider theme={theme}>
      <AuthProvider
        {...oidcConfig}
        onSigninCallback={async (user) => {
          console.log('User signed in', user);
          window.location.href = '/groups';
        }}
      >
        <App />
      </AuthProvider>
    </ChakraProvider>
  );
};

const container = document.getElementById('root') as HTMLElement;
const root = createRoot(container);
root.render(<Index />);
