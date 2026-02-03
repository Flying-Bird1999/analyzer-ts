// Button 组件实现
export interface ButtonProps {
  label: string;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'danger';
  loading?: boolean;  // 新增：加载状态
}

const Button: React.FC<ButtonProps> = ({ label, onClick, variant = 'primary', loading = false }) => {
  return (
    <button
      className={`btn btn-${variant} ${loading ? 'btn-loading' : ''}`}
      onClick={onClick}
      disabled={loading}
    >
      {loading ? 'Loading...' : label}
    </button>
  );
};

export const IconButton: React.FC<{ icon: string; onClick?: () => void; title?: string }> = ({ icon, onClick, title }) => {
  return <button className="btn-icon" onClick={onClick} title={title}>{icon}</button>;
};

export const LinkButton: React.FC<{ label: string; href?: string; onClick?: () => void }> = ({ label, href, onClick }) => {
  if (href) {
    return <a href={href} className="btn-link">{label}</a>;
  }
  return <button className="btn-link" onClick={onClick}>{label}</button>;
};

export default Button;