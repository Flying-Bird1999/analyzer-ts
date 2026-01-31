// Button 组件实现
// export interface ButtonProps {
//   label: string;
//   onClick?: () => void;
// }

export const Button: React.FC<{ label: string; onClick?: () => void }> = ({ label, onClick }) => {
  return <button onClick={onClick}>{label}</button>;
};
