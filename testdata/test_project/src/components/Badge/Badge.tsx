// Badge 组件实现
import { Input } from '../Input/Input';

export interface BadgeProps {
  text: string;
  count?: number;
  variant?: 'default' | 'success' | 'warning' | 'error';
}

export const Badge: React.FC<BadgeProps> = ({ text, count, variant = 'default' }) => {
  return (
    <span className={`badge badge-${variant}`}>
      {text}
      {count !== undefined && <span className="badge-count">{count}</span>}
    </span>
  );
};

// Badge 组件引用 Input 组件（增加依赖复杂度）
export const BadgeWithInput: React.FC<BadgeProps & { placeholder?: string }> = (props) => {
  return (
    <div className="badge-with-input">
      <Badge {...props} />
      <Input value="" placeholder={props.placeholder} />
    </div>
  );
};
