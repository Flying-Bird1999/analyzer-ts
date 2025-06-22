

/** 广播 state类型定义 */
export type BroadcastDataType = {
  name: string;
  sendTargets: ESendTarget;
  fansPages?: { label: string; value: string };
  platformChannelName?: string;
  contactTags?: string[];
  subTemplate?: { label: string; value: string }[];
  sendExpectCount?: number;
  contents: ContentType[];
  bcType?: BroadcastSendType;
  sendTime?: Dayjs | string;
  customGroup?: ICustomGroupDetail | null;
  sendStatus?: ESendStatus;
  sendSucTargetCount?: number;
  sendTargetCount?: number;
};


/** 广播联络人类型枚举 */
export enum ESendTarget {
  /** 24小时内互动过的Facebook联络人 */
  ALL_24_HOUR_FB_CONTACT = 'ALL_24_HOUR',
  /** 订阅中的Facebook联络人 */
  SUB_FB_CONTACT = 'SUBSCRIBING',
  /** 指定顾客分群 */
  CUSTOM_GROUP = 'CUSTOM_GROUP_SUBED',
  /** 指定标签的Facebook联络人 */
  TAG_WA_CONTACT = 'TAG_SUBED',
}


/** 跟进模版详情 */
export type IAuthTemplateDetail = {
  payload?: {
    title?: string;
    subtitle?: string;
    text?: string;
    url?: string;
    attachmentId?: string;
    imageUrl?: string;
    elements?: {
      title?: string;
      subtitle?: string;
      imageUrl?: string;
      buttons?: {
        type?: string;
        url?: string;
        title?: string;
      }[];
    }[];
    buttons?: {
      type?: string;
      url?: string;
      title?: string;
    }[];
  };
  templateStructureType?: ETemplateStructureType;
};


/** 广播消息类型 */
export enum ContentMessageType {
  /** 文字消息 */
  words = 'TEXT',
  /** 图片消息 */
  picture = 'IMAGE',
  /** 图文消息 */
  words_picture = 'TEXT_IMAGE',
  /** 消息模版 */
  template = 'TEMPLATE',
  /** 卡片消息 */
  card = 'FB_GENERIC_TEMPLATE',
  /** 授权订阅消息 */
  subscription = 'FB_MARKETING_TEMPLATE',
}


export enum ComponentType {
  // 页头
  Header = 'HEADER',
  // 正文
  Body = 'BODY',
  // 页脚
  Footer = 'FOOTER',
  // 按钮
  Buttons = 'BUTTONS',
  // 卡片
  CAROUSEL = 'CAROUSEL',
}


export enum DetailButtonsType {
  /** 快捷回复 */
  quick_reply = 'QUICK_REPLY',
  /** PHONE_NUMBER */
  phone_number = 'PHONE_NUMBER',
  /** URL */
  url = 'URL',
}


export enum QuickReplyButtonEffect {
  /** 无 */
  none = '',
  /** 订阅商家讯息 */
  subscribe_information = '{{data.subscribe}}',
  /** 退订商家讯息 */
  unsubscribe_information = '{{data.unsubscribe}}',
}


/** 广播内容 */
export type ContentType = {
  /** 消息类型 */
  contentType: ContentMessageType;
  /** 图片 */
  attachmentUrl: string;
  /** 文字 */
  text: string;
  /** whatsapp制定模版 */
  template?: {
    name?: string;
    customKey?: string;
    businessKey?: string;
    templateType?: TemplateDetailType;
    lang?: LanguageData;
    auditStatus?: string;
    score?: string;
    content: WhatsappTemplateComponentModel[];
    customParam?: Record<string, string>;
  };
  /** fb/ig 授权模板信息 */
  fbMarketingTemplate?: IAuthTemplateItem;
  /** fb/ig 卡片 */
  fbGenericTemplates?: ICardType[];
};
/** 授权模版类型定义 */
export enum ETemplateStructureType {
  /** 图片 */
  IMAGE = 'IMAGE',
  /** 文本 */
  PLAIN_TEXT = 'PLAIN_TEXT',
  /** 图文 */
  IMAGE_TEXT = 'TEXT_IMAGE',
  /** 卡片 */
  CARD = 'CARD',
}


/** 模版是否订阅 */
export enum ESubscribedStatus {
  /** 订阅 */
  SUBSCRIPTED = 1,
  /** 未曾订阅 */
  NO_SUBSCRIPT = 2,
  /** 取消订阅（曾经订阅过） */
  CANCEL_SUBSCRIPT = 0,
}


export interface WhatsappTemplateComponentModel {
  /**类型
BODY、HEADER、FOOTER 和 BUTTONS */
  type?: ComponentType;
  /**文本内容 */
  text?: string;
  /**当type=HEADER时,有以下分类
类型(TEXT, IMAGE, DOCUMENT, VIDEO) */
  format?: MessageTitleType;
  /**type=HEADER时且format!=TEXT时,header的具体内容 */
  example?: {
    header_handle?: string[];
    header_text?: string[];
    body_text?: string[][];
  };
  /**按钮列表 */
  buttons?: ButtonRespVO[];
  /** 卡片列表 */
  cards?: {
    components: WhatsappTemplateComponentModel[];
  }[];
}


export interface ButtonRespVO {
  /**类型(PHONE_NUMBER、URL 和 QUICK_REPLY) */
  type: DetailButtonsType;
  /**按钮的文本显示内容 */
  text: string;
  /**当type=URL时按钮,填写跳转地址 */
  url?: string;
  /**当type=QUICK_REPLY,可填写回传值 */
  payload?: QuickReplyButtonEffect | string;
  /**电话 */
  phone?: string;
  /**国家 */
  country?: string;
  /**带国家编码,如+8618814098372 */
  phone_number?: string;
}


/** 广播发送时间类型 */
export enum BroadcastSendType {
  /** 立即发送 */
  NOW_SEND = 'REAL_TIME',
  /** 指定时间发送 */
  SPECIFY_SEND = 'TIMING',
}


/** 广播发送状态 */
export enum ESendStatus {
  /** 发送成功 */
  SUCCESS = 'SUCCESS',
  /** 发送中 */
  SENDING = 'SENDING',
  /** 发送失败 */
  FAILED = 'FAILED',
  /** 部分发送成功 */
  PART_SUCCESS = 'PART_SUCCESS',
  /** 定时广播--等待发送 */
  PENDING = 'PENDING',
  /** 等待再次发送 */
  WAITING = 'WAITING',
}


/** 模版列表 */
export type IAuthTemplateItem = {
  id?: string;
  name: string;
  payload: IPayload;
  postback: IAuthTemplateDetail[];
  messageCount?: number;
  subscribedStatus?: ESubscribedStatus;
};


/** 授权模版详情 */
export type IPayload = {
  title: string;
  imageUrl: string;
  payload?: string;
};


export enum TemplateDetailType {
  /** 营销推广 */
  marketing_promotion = 'MARKETING',
  /** 客服消息 */
  customer_service_information = 'TRANSACTIONAL',
  /** 飞鸽新的枚举值 --- 客服消息 */
  utility = 'UTILITY',
}
/** 模版语言 */
export enum LanguageData {
  /** 简体中文 */
  ZH_CN = 'zh-hans-cn',
  /** 繁体中文 */
  ZH_TW = 'zh-hant-hk',
  /** 英语 */
  EN = 'en',
  /** 越南语 */
  VN = 'vi',
  /** 泰语 */
  TH = 'th',
}


export enum MessageTitleType {
  /** 无 */
  none = '',
  /** 文本 */
  text = 'TEXT',
  /** 图片 */
  image = 'IMAGE',
  /** 视频 */
  video = 'VIDEO',
  /** 文档 */
  file = 'DOCUMENT',
}


/** 卡片消息类型 */
export type ICardType = {
  title?: string;
  subtitle?: string;
  imageUrl?: string;
  buttons: {
    type: string;
    url: string;
    title: string;
  }[];
};


/** wa选择指定分群详情 */
export interface ICustomGroupDetail {
  id?: string;
  /** 分群名称 */
  name?: string;
  /** 分群顾客人数 */
  count?: number;
  /** 创建时间 */
  created_at?: string;
  /** 发送范围 */
  optin?: WaSendTarget;
}


/** wa发送对象枚举 */
export enum WaSendTarget {
  /** 全部WhatsApp联络人 */
  ALL_WA_CONTACT = 'ALL_WA_CONTACT',
  /** 24小时内互动过的WhatsApp联络人 */
  ALL_24_HOUR_WA_CONTACT = 'ALL_24_HOUR_WA_CONTACT',
  /** 指定WhatsApp联络人 */
  PART_WA_CONTACT = 'PART_WA_CONTACT',
  /** 指定顾客分群 */
  CUSTOM_GROUP = 'CUSTOM_GROUP',
  /** 指定顾客分群 --- 所有发送对象 */
  CUSTOM_GROUP_HAS_PHONE = 'CUSTOM_GROUP_HAS_PHONE',
  /** 指定顾客分群 --- 订阅中的发送对象 */
  CUSTOM_GROUP_SUBED = 'CUSTOM_GROUP_SUBED',
  /** 订阅中的WhatsApp联络人 */
  SUB_WA_CONTACT = 'SUB_WA_CONTACT',
  /** 暂未订阅的WhatsApp联络人 */
  NOT_SUB_WA_CONTACT = 'NOT_SUB_WA_CONTACT',
  /** 指定标签的WhatsApp联络人 */
  TAG_WA_CONTACT = 'TAG_WA_CONTACT',
  // 兼容旧数据的展示
  ALL_24_HOUR = 'ALL_24_HOUR',
  PART_24_HOUR = 'PART_24_HOUR',
}
