import type {
  AccountCredentialStatus,
  UpstreamAPIStatus,
  UpstreamCheckinStatus,
  UpstreamCredentialKind,
  UpstreamPlatformKind,
} from '../../types';

export const platformOptions: Array<{ label: string; value: UpstreamPlatformKind }> = [
  { label: 'New API', value: 'new_api' },
  { label: 'Sub2API', value: 'sub2api' },
];

export const platformLabels: Record<UpstreamPlatformKind, string> = {
  new_api: 'New API',
  sub2api: 'Sub2API',
};

export const credentialKindOptions: Array<{ label: string; value: UpstreamCredentialKind }> = [
  { label: '不配置', value: 'none' },
  { label: 'Cookie / Session Header', value: 'cookie' },
  { label: 'Access Token', value: 'access_token' },
  { label: 'Refresh Token', value: 'refresh_token' },
  { label: 'JSON 凭据', value: 'json' },
];

export const apiStatusOptions: Array<{ label: string; value: UpstreamAPIStatus | 'all' }> = [
  { label: 'API 状态：全部', value: 'all' },
  { label: '健康', value: 'healthy' },
  { label: '警告', value: 'warning' },
  { label: '失败', value: 'failed' },
  { label: '禁用', value: 'disabled' },
  { label: '未知', value: 'unknown' },
];

export const accountStatusOptions: Array<{ label: string; value: AccountCredentialStatus | 'all' }> = [
  { label: '账号凭据：全部', value: 'all' },
  { label: '未配置', value: 'not_configured' },
  { label: '有效', value: 'valid' },
  { label: '过期', value: 'expired' },
  { label: '失败', value: 'failed' },
  { label: '需人工处理', value: 'action_required' },
];

export const checkinStatusOptions: Array<{ label: string; value: UpstreamCheckinStatus | 'all' }> = [
  { label: '签到：全部', value: 'all' },
  { label: '不支持', value: 'unsupported' },
  { label: '未配置', value: 'not_configured' },
  { label: '已签到', value: 'checked' },
  { label: '未签到', value: 'unchecked' },
  { label: '失败', value: 'failed' },
  { label: '需人工处理', value: 'action_required' },
];
