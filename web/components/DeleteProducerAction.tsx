"use client";

import DeleteButton from "./DeleteButton";
import { deleteProducerAction } from "@/app/actions/data";

export default function DeleteProducerAction({ producerId }: { producerId: string }) {
  return (
    <DeleteButton
      onDelete={() => deleteProducerAction(producerId)}
      redirectTo="/producers"
      confirmMessage="Atenção: excluir este produtor também excluirá todos os seus talhões e alertas associados. Deseja prosseguir?"
      label="Excluir Produtor"
    />
  );
}

