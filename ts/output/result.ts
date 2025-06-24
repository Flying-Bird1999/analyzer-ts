

export interface ProductModel extends IProductQuantityTypesCamelCase {
  id: string;
  retailStatus: string;
  /** 商品名称 */
  titleTranslations: TranslationModel;
  /** 原价 */
  price: PriceModel;
  /** 特价 */
  priceSale: PriceModel;
  /** 规格商品中的最低价 */
  lowestPrice: PriceModel;
  /** 特价规格商品中的最低价 */
  lowestPriceSale: PriceModel;
  /** 隐藏价格 */
  hidePrice: boolean;
  /** 规格商品价格是否同主商品 */
  samePrice: boolean;
  /** 成本价 */
  cost: PriceModel;
  /** 会员价 */
  memberPrice: PriceModel;
  flashPriceSets: FlashPriceSet[];
  /** 是否预购 */
  isPreorder: boolean;
  /** 预购讯息 */
  preorderNoteTranslations: TranslationModel;
  /** 重量 */
  weight: number;
  /** 商品目前数量 */
  quantity: number;
  /** 商品是否无限数量 */
  unlimitedQuantity: boolean;
  /** 图片 */
  medias: MediasItem[];
  /** 更多图片 */
  detailMedias: unknown[];
  /** 所属分类的id */
  categoryIds: Array<string>;
  fieldTitles: FieldTitles[];
  /** 供应商 */
  // supplier?: string;
  // sku?: string;
  // /** 商品条码 */
  barcode?: string;
  /** 规格名称 */
  /** 多规格数据 */
  variations: Variation[];
  /** 商品规格 */
  variantOptions: VariantOptions[];
  /** 商品分类 */
  categories: CategoryModel;
  /** 储位编号 */
  // locationId?: string;
  /** 商品摘要 */
  summaryTranslations: TranslationModel;
  /** 商品描述 */
  descriptionTranslations: TranslationModel;
  /** SEO标题 */
  seoTitleTranslations: TranslationModel;
  /** SEO描述 */
  seoDescriptionTranslations: TranslationModel;
  /** SEO关键字 */
  // seoKeywords: string;
  link: string;
  /** 商品缺货是否提醒 */
  isReminderActive: boolean;
  /** 显示相关商品 */
  showCustomRelatedProducts: boolean;
  /** 标签 */
  tags: string[];
  /** 排除的送货方式 */
  blacklistedDeliveryOptionIds: string[];
  /** 排除的付款方式 */
  blacklistedPaymentIds: string[];
  /** 单次购买上限 */
  maxOrderQuantity: number;
  barcodeType: string;
  updatedAt: string;
  createdAt: string;
  createdBy: string;
  isExcludedPromotion: boolean;
  taxable: boolean;
  /** 预设商品销售开始时间 */
  availableStartTime?: string;
  /** 预设商品销售结束时间 */
  availableEndTime?: string;
  // schedulePublishAt?: string;
  labels: unknown[];
  productPriceTiers: ProductPriceTiers[];
  /** 锁库存数量 */
  lockedInventoryCount: number;
  /** 商品状态 active上架 draft下架 hide隐藏 */
  status: ProductStatusType;
  /** 关键字是否生效 true失效 false生效 */
  isKeywordNotEffect: boolean;
  // [key:string]: any
  exclusivePriceInfo?: ExclusivePriceInfo;
  /** 未压缩商品图片 */
  originalUrl?: string;
}


type VariantOptions = {
  id: string;
  type: string;
  nameTranslations: TranslationModel;
  index: number;
};


export type Variation = {
  id: string;
  /** 规格名称 */
  fieldsTranslations: TranslationModel;
  /** 原价格 */
  price: PriceModel;
  /** 特价 */
  priceSale: PriceModel;
  /** 会员价 */
  memberPrice: PriceModel;
  /** 成本价 */
  cost: PriceModel;
  /** 规格现有数量 */
  quantity: number;
  /** 规格是否无限数量 */
  // unlimitedQuantity: boolean;
  /** 图片 */
  // media: unknown[];
  sku?: string;
  /** 储位编号 */
  // locationId: string;
  gtin: string;
  /** 对应规格类别ID */
  variantOptionIds: string[];
  /** 条码编号 */
  barcode?: string;
  /** 条码编号类别 */
  barcodeType: string;
  weight: number;
  /** 锁库存 */
  lockedInventoryCount: number;
} & IProductQuantityTypesCamelCase;


export type CategoryModel = {
  id: string;
  /** 分类名称 */
  nameTranslations?: TranslationModel;
  /** SEO标题名称 */
  seoTitleTranslations?: TranslationModel;
  /** SEO简介 */
  seoDescriptionTranslations?: TranslationModel;
  /** SEO关键字 */
  seoKeywords?: string;
  key?: string;
  /** 当前状态 */
  status?: 'active' | 'removed';
  /** 分类横图 */
  bannerMedias?: unknown[];
  /** 母分类ID */
  parentId?: string;
  /** 分类权重 */
  priority?: number;
  createdBy?: 'admin' | 'pos' | 'openapi';
  /** 子分类 */
  children?: CategoryModel[];
};

/** 源数据 */
export interface SkuProductItem {
  /** sku 主键id */
  id: number | string;
  salesProductId: number;
  /** sku id - 很长那个 */
  skuId: string;
  /** 默认关键字 */
  defaultKey: string;
  /** 定制关键字 */
  customKeys: string[];
  /** 关键字展示 - 后端拼接好的 */
  customKeysLabelStr: string;
  customKeysLabel: string[];
  /** 锁库存数量 */
  lockedInventoryCount: number;
  /** 商品状态 active上架 draft下架 hide隐藏 */
  status: ProductStatusType;
  /** 最终需要展示的价格（后端计算好的展示价格，包含flash price） */
  price: PriceModel;
  priceSale: PriceModel;
}


export type FlashPriceSet = {
  endAt: string;
  id: string;
  priceSet: PriceSet;
  startAt: string;
  updatedAt: string;
};


export type PriceModel = {
  cents: number;
  currencySymbol: string;
  currencyIso: string;
  label: string;
  dollars: number;
};


type ProductStatusType = 'active' | 'draft' | 'hide' | 'hidden';


export interface ExclusivePriceSkuItem {
  skuId: string;
  priceCents: number;
  priceDollars: number;
}


type ImagesType =
  | 'original'
  | 'faviconSmall'
  | 'favicon'
  | 'faviconLarge'
  | 'transparentThumb'
  | 'transparentLarge'
  | 'transparentXlarge'
  | 'thumb'
  | 'source';


type FieldTitles = {
  key: string;
  label: string;
  nameTranslations: TranslationModel;
  index: number;
};


export enum ProductLiveStatus {
  INTRODUCING = 'INTRODUCING',
  DISPLAYING = 'DISPLAYING',
  HIDDEN = 'HIDDEN',
}


export enum EStatus {
  /** 未开始 */
  PLAN = 'PLAN',
  /** 进行中 */
  IN_PROGRESS = 'IN_PROGRESS',
  /** 已结束 */
  END = 'END',
}


export interface HostPanelProductRes {
  // 正在推荐的商品 (直播间商品列表的结构)
  recommendProduct: LiveProductItem | null;
  // 待推荐的商品 (直播间商品列表的结构)
  loadRecommendProduct: LiveProductItem | null;
}

/** 源数据 */
export interface LiveProductItem {
  /** 主键Id */
  id: number | string;
  /** 直播贴文id */
  salesId: number | string;
  /** spuId */
  spuId: string;
  /** 关键字生效状态(true=生效,false=不生效) */
  keywordStatus: boolean;
  /** 系统默认关键字 */
  defaultKey: string;
  /** 商品信息 */
  productInfo: ProductModel;
  /** sku信息 */
  skuProductList: SkuProductItem[];
  /** 商品最后推荐的时间 */
  recommendedTime: string;
  /** 商品编号 */
  customNumber: string;
  /** 多个商品编号集合 */
  customNumbers: string[];
  /** 最终需要展示的价格（后端计算好的展示价格，包含flash price） */
  price: PriceModel;
  priceSale: PriceModel;
  /** 买家信息 商品加购数量映射 */
  countCommentSkuMap: { [key: string]: number };
  /** 压缩过的商品图片 */
  compressedPicture?: string;
  /** 是否在推荐中的商品 */
  currentRecommend?: boolean;
  /** 商品状态 */
  productStatus: ProductLiveStatus;
  /** 是否 PING */
  currentPinning?: boolean;
  /** 商品限定价 */
  exclusiveSpuEvent?: IExclusiveSpuEvent;
  /** 商品类型 */
  productType: string;
}


type ProductPriceTiers = {
  _id: string;
  createdAt: string;
  membershipTierId: string;
  productId: string;
  status: string;
  updatedAt: string;
};


export interface IExclusiveSpuEvent {
  startTime: string;
  endTime: string;
  eventStatus: EStatus;
  eventType: string;
  id: number;
  merchantId: string;
  spuId: string;
  version: number;
  eventDetailList: IEventDetailList[];
}


export type TranslationModel = {
  [key in SupportedLanguages]?: string[] | string;
};


type PriceSet = {
  flashPriceCampaignId: string;
  id: string;
  price: IPriceModel;
  priceDetails: [];
  priceSale: IPriceModel;
  status: string;
  type: string;
};


export interface ExclusivePriceInfo {
  id: string;
  spuId: string;
  startTime: string;
  endTime: string;
  status: ExclusivePriceStatus;
  skuPriceList: ExclusivePriceSkuItem[];
}


export enum ExclusivePriceStatus {
  NOT_STARTED = 0,
  ING = 1,
  ENDED = 2,
}


export type MediasItem = {
  images: {
    [key in ImagesType]: ImagesAttribute;
  };
  _id: string;
};


type ImagesAttribute = {
  width: number;
  height: number;
  url: string;
};


export interface IEventDetailList {
  cents: number;
  eventId: number;
  id: number;
  merchantId: string;
  price: number;
  skuId: string;
  spuId: string;
}
