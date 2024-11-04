import packageJson from '../../package.json';

export const PORTAL_VERSION = packageJson.version;
export let CLIENT_ID:string;
export let BACKEND_URL:string;
export let APP_URL:string;

// if (!process.env.NODE_ENV || process.env.NODE_ENV === 'development') {
//     CLIENT_ID = 'veda-dafsxhsztbsczrmmbftw-10000000';
//     BACKEND_URL = 'http://localhost:8081';
//     APP_URL = 'http://localhost:5173'
// } 
// else {
    // Client ID goes here
    CLIENT_ID = 'custos-kgap8hu6ih4hddvlzzlb-10000000';
    BACKEND_URL = 'https://api.playground.usecustos.org';
    APP_URL = 'http://localhost:5173'
// }

export const APP_REDIRECT_URI = `${APP_URL}/oauth-callback`;
export const TENANT_ID = '10000000';