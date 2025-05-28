

export interface CurrentStaff {
  readonly id: string;
  readonly name: string;
  readonly email: string;
  readonly merchant_ids: string[];
  readonly organization_ids: string[];
  readonly role_keys: string[];
  readonly locale_code: string;
}


export type SupportedLanguages = _SupportedLanguages;
export type _SupportedLanguages = 'en' | 'zh-cn' | 'zh-hant' | 'vi' | 'th';


export default interface PreloadedState {
  currentMerchant: MerchantData;
  currentStaff: CurrentStaff;
  currentStaffPermission: { [key: string]: string[] };
  locale: SupportedLanguages;
  callBackData: { [key: string]: any };
}


export interface MerchantData {
  readonly id: string;
  readonly email: string;
  readonly owner_id: string;
  readonly staff_id: string;
  readonly logo_media_url: string;
  readonly handle: string;
  readonly admin_url: string;
  readonly url: string;
  readonly rollout_keys: string[];
  readonly rolloutKeys: string[];
  readonly supported_languages: LanguageKey[];
  readonly name: string;
  readonly base_country_code: string;
  readonly title_translations: {
    [key in SupportedLanguages]: string;
  };
  readonly subtitle_translations: {
    [key in SupportedLanguages]: string;
  };
  readonly icon?: string;
  readonly base_currency?: BaseCurrency;
  readonly created_at: string;
  readonly current_plan_key: string;
  readonly base_currency_code: string;
  readonly raw_subscription_currency: string;
  readonly subscription_currency: BaseCurrency;
  readonly default_language_code: string;
  readonly sl_payment_merchant_id: string;
  readonly brand_home_url: string;
}


export type LanguageKey =
  | 'en'
  | 'zh-hant'
  | 'zh-hk'
  | 'zh-tw'
  | 'zh-cn'
  | 'vi'
  | 'ms'
  | 'ja'
  | 'th'
  | 'id'
  | 'de'
  | 'fr';


export interface BaseCurrency {
  alternate_symbol: string;
  iso_code: string;
  symbol_first: boolean;
}
