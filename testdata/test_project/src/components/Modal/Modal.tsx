// Modal 组件实现
import { useEffect, useRef, useState } from 'react';
import { Button } from '../Button/Button';
import logoImage from '../../assets/logo.png';
import '../../assets/modal.css';

export interface ModalProps {
  isOpen: boolean;
  title?: string;
  onClose?: () => void;
  children: React.ReactNode;
}

export const Modal: React.FC<ModalProps> = ({ isOpen, title, onClose, children }) => {
  const overlayRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose?.();
      }
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div ref={overlayRef} className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        {title && <div className="modal-header"><img src={logoImage} alt="logo" className="modal-logo" /><h2>{title}</h2></div>}
        <div className="modal-body">{children}</div>
        <div className="modal-footer">
          <Button label="Close" onClick={onClose} />
        </div>
      </div>
    </div>
  );
};

export const useModal = () => {
  const [isOpen, setIsOpen] = useState(false);
  const open = () => setIsOpen(true);
  const close = () => setIsOpen(false);
  const toggle = () => setIsOpen(!isOpen);
  return { isOpen, open, close, toggle };
};
