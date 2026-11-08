"use client";

import DeleteButton from "./DeleteButton";
import { deletePropertyAction } from "@/app/actions/data";

export default function DeletePropertyAction({ propertyId }: { propertyId: string }) {
  return (
    <DeleteButton
      onDelete={() => deletePropertyAction(propertyId)}
      redirectTo="/"
      confirmMessage="Tem certeza que deseja excluir esta propriedade e seu histórico de alertas?"
      label="Excluir Talhão"
    />
  );
}

