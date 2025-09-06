import { APP_NAME } from '@constants/app';

function formatAppInfo(appName: string, appVersion: string): string {
    return `${appName} v${appVersion}`;
}

export function getAppInfo(): string {
    return formatAppInfo(APP_NAME, '1.0.0');
}
