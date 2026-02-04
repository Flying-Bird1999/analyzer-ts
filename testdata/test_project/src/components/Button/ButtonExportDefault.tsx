// Button 组件实现 - 用于测试 export default () => {} 场景
export interface ButtonProps {
  label: string;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'danger';
  loading?: boolean;
}

// export default 箭头函数（用于测试 export default 场景）
export default () => {
  return <button>Click</button>
};
