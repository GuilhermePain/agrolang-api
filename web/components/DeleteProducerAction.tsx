"use client";

import DeleteButton from "./DeleteButton";
import { deleteProducer } from "@/lib/api";

export default function DeleteProducerAction({ producerId }: { producerId: string }) {
  return (
    <DeleteButton
      onDelete={() => deleteProducer(producerId)}
      redirectTo="/producers"
      confirmMessage="Atenção: excluir este produtor também excluirá todos os seus talhões e alertas associados. Deseja prosseguir?"
      label="Excluir Produtor"
    />
  );
}

