enum PolicyDimensionEnum {
  STORE = 'STORE',
  SUB = 'SUB',
}

export interface CurrentRes {
  /**店铺ID */
  storeId: string;
  /**佣金政策维度，SUB - 基于已购套餐的佣金政策，STORE - 基于店铺的佣金政策 */
  policyDimension: PolicyDimensionEnum;
  /**万分位抽佣比例，例如0.05%这里是：5 */
  scale: number;
  /**剩余免佣额度,单位厘 */
  exemptQuota: number;
  /**免佣额度货币类型 */
  exemptCurrency: string;
}

interface StoreDetailReq {
  /**店铺ID */
  storeId: string;
  policyType: number;
}

interface StoreCommissionScaleVO {
  /**套餐ID，子政策类型为基于将来的套餐才会有该字段值 */
  planId: string;
  /**套餐名称（含多语言），子政策类型为基于将来的套餐才会有该字段值 */
  planName?: Record<string, string>;
  /**万分位抽佣比例，例如0.05%这里是：5 */
  scale: number;
  /**剩余免佣额度 */
  exemptQuota: number;
  /**免佣额度货币类型 */
  exemptCurrency: string;
}

interface StoreDetailRes {
  /**店铺ID */
  storeId: string;
  /**子政策有效性，true:有效，false:无效 */
  valid: boolean;
  /**子政策类型，0：基于固定时段，1：基于将来的套餐版本 */
  policyType: number;
  /**子政策内容 */
  commissionScales: StoreCommissionScaleVO[];
  /**开始时间 */
  startTime: number;
  /**结束时间 */
  endTime: number;
}

export {
  StoreDetailReq,
  StoreDetailRes,
}