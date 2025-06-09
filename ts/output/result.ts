

interface Name extends ISelectProductSetItem {
  name: string;
  value: string;
}

/** 选中组合商品 */
export type ISelectProductSetItem = {
    id: string;
    /** 组合商品图片 */
    image: string;
    /** 组合商品名称 */
    name: string;
    /** 价格 */
    price: {
        /** 原价 */
        origin?: string;
        /** 销售价 */
        sales: string;
    };
    /**
     * 是否所有的子商品含有多规格商品
     * 排除掉 无规格/单规格
     */
    isChildProductHasMoreVariations: boolean;
    /** 组合列表 */
    subItems: ISubItems[];
    /** defaultKey */
    defaultKey?: string;
};

/** 组合类型 */
export type ISubItems = {
    /** mock 组合id */
    id: string;
    selectedList: ISelectItem[];
    /** 预设 keywords */
    keyword?: string[];
};

/** 回调选中数据类型 */
export type ISelectItem = {
    spuId: string;
    skuId?: string;
    quantity: number;
    selectedLabel: string;
};
