import React, { useContext, useState } from "react";
import { DialogProps } from "@radix-ui/react-dialog";

interface DialogContextType {
  create: (node: React.FC<DialogProps>) => void;
}

const DialogContext = React.createContext<DialogContextType | null>(null);

export const DialogProvider = ({ children }: { children: React.ReactNode }) => {
  const [dialogs, setDialogs] = useState<React.FC<DialogProps>[]>([]);

  const create = (dialog: React.FC<DialogProps>) =>
    setDialogs([...dialogs, dialog]);

  return (
    <DialogContext.Provider
      value={{
        create,
      }}
    >
      {children}
      {dialogs.map((Dialog, i) => (
        <Dialog
          key={i}
          open
          onOpenChange={(value) => {
            if (!value) {
              setDialogs(dialogs.filter((_, index) => index !== i));
            }
          }}
        />
      ))}
    </DialogContext.Provider>
  );
};

export const useDialog = () => {
  const context = useContext(DialogContext);
  if (!context) throw new Error("can only useDialog in DialogProvider");
  return context;
};
