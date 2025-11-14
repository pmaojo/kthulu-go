import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PropertiesPanel } from "@/components/PropertiesPanel";
import { ElementProperties } from "@/types/properties";

describe("PropertiesPanel", () => {
  const baseElement: ElementProperties = {
    id: "service-1",
    type: "service",
    name: "Servicio Principal",
    description: "Servicio base",
    status: "active",
  };

  it("calls onApply with updated data", async () => {
    const handleApply = vi.fn();
    const user = userEvent.setup();

    render(
      <PropertiesPanel
        isOpen
        onClose={() => {}}
        selectedElement={baseElement}
        onApply={handleApply}
      />
    );

    const nameInput = screen.getByPlaceholderText("Nombre del elemento") as HTMLInputElement;
    await user.clear(nameInput);
    await user.type(nameInput, "Servicio Actualizado");

    await user.click(screen.getByRole("button", { name: /aplicar/i }));

    expect(handleApply).toHaveBeenCalledTimes(1);
    expect(handleApply).toHaveBeenCalledWith(
      expect.objectContaining({
        id: baseElement.id,
        type: "service",
        name: "Servicio Actualizado",
        description: baseElement.description,
        status: "active",
      })
    );
  });

  it("calls onDelete with the element id", async () => {
    const handleDelete = vi.fn();
    const user = userEvent.setup();

    render(
      <PropertiesPanel
        isOpen
        onClose={() => {}}
        selectedElement={baseElement}
        onDelete={handleDelete}
      />
    );

    await user.click(screen.getByRole("button", { name: /eliminar/i }));

    expect(handleDelete).toHaveBeenCalledTimes(1);
    expect(handleDelete).toHaveBeenCalledWith(baseElement.id);
  });

  it("disables action buttons when callbacks are missing", () => {
    render(
      <PropertiesPanel
        isOpen
        onClose={() => {}}
        selectedElement={baseElement}
      />
    );

    expect(screen.getByRole("button", { name: /aplicar/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /eliminar/i })).toBeDisabled();
  });
});
