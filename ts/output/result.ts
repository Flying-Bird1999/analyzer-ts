

/**
 * 打印移除排除项
 */

export enum EnumPrintExcludes {
  /** 标签id */
  TagId = 'tag_id',
  /** 电话 */
  Phone = 'phone',
  /** 地址 */
  Address = 'address',
  /** 留言时间 */
  CommentTime = 'comment_time',
  /** 价格 */
  Price = 'price',
  /** 总计 */
  Total = 'total',
  /** 活动名称 */
  ActivityTitle = 'activity_title',
}

export declare type Env = 'DEV' | 'TEST' | 'FAT' | 'PRO';


/**
 * 打印相关配置
 */
export type TPrintConfigs = TWithMerchantConfigs<IMerchantsCustomizePrintConfigs>;


/**
 * 商家白名单自定义配置
 */
export interface IMerchantsCustomizePrintConfigs {
  /**打印尺寸 */
  size: {
    /**限制打印内容的宽 */
    width: number;
    /**限制打印内容的高 */
    height: number;
  };
}


export interface ConfigCenterType {
  // ec地址
  ec_admin_url: string;
  // 是否开启sentry
  enable_sentry: string;
  // chatroom环境
  chatroom_env: string;
  // 多语言语言环境
  lang_env: string;
  // java地址
  java_api_url: string;
  // bff地址（旧）
  bff_api_url: string;
  // bff地址（新）
  bff_api_url2: string;
  // 无痕埋点环境
  slq_env: Env;
  // IM环境
  im_env: string;
  // IM签名私钥
  im_signPrv: string;
  // 回放重放过期时间
  replay_expiration_time: string;
  // 赠品的默认的媒体ID
  gift_default_media_id: string;
  // viewer端访问域名
  sc_live_domain: string;
  // hummer订阅的domain，暂时仅提供给压测环境需要这个配置
  hummer_ws_domain: string;
  /**
   * 打印白名单配置
   */
  print?: TPrintConfigs;
  /**
   * 打印自定义排除项
   */
  print_excludes: TPrintContentExcludesConfigs;
  /** sentry自定义忽略规则 */
  sentry_ignore_errors: string[];
  // 登录鉴权相关接口切换开关
  login_api_switch: string;
}


/**
 * 打印内容自定义配置
 */
export type TPrintContentExcludesConfigs = TWithMerchantConfigs<EnumPrintExcludes[]>;


/**
 * 白名单用户包裹类型
 */
export type TWithMerchantConfigs<T> = Record<IMerchantId, T>;
/**
 * merchantId 商家唯一标识
 */
type IMerchantId = string;
