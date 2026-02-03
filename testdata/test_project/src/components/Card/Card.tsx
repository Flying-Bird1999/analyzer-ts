// Card 组件实现
import Button from '../Button/Button';
import { ThemeColor, ButtonSize, Direction } from '../../types/enums';

export interface CardProps {
  title: string;
  description?: string;
  footer?: React.ReactNode;
  theme?: ThemeColor;    // 使用枚举
  size?: ButtonSize;      // 使用枚举
}

const Card: React.FC<CardProps> = ({ title, description, footer, theme = ThemeColor.Primary, size = ButtonSize.Medium }) => {
  return (
    <div className={`card card-${theme} card-${size}`}>
      <div className="card-header">
        <h3>{title}</h3>
      </div>
      {description && <div className="card-body">{description}</div>}
      {footer && <div className="card-footer">{footer}</div>}
    </div>
  );
};

// Card 组件引用 Button 组件和枚举
export const CardWithButton: React.FC<CardProps & { buttonText?: string; direction?: Direction }> = (props) => {
  return (
    <Card {...props}>
      <Button label={props.buttonText || "Action"} onClick={() => {}} />
    </Card>
  );
};

export default Card;