"use client";

import { useEffect, useState, type FormEvent } from "react";
import { Badge, Button, Card, Container, Input, Text } from "@movoz/ui-web";
import { api } from "@/lib/api";
import type { MainAttribute, SubAttribute } from "@/lib/types";

export default function AttributesPage() {
  const [mainAttributes, setMainAttributes] = useState<MainAttribute[]>([]);
  const [subAttributesByMain, setSubAttributesByMain] = useState<Record<number, SubAttribute[]>>({});
  const [newMainKey, setNewMainKey] = useState("");
  const [newMainName, setNewMainName] = useState("");
  const [newSubName, setNewSubName] = useState<Record<number, string>>({});

  async function load() {
    const mains = await api.get<MainAttribute[]>("/api/main-attributes");
    setMainAttributes(mains);
    const entries = await Promise.all(
      mains.map(async (m) => [m.id, await api.get<SubAttribute[]>(`/api/sub-attributes?main_attribute_id=${m.id}&active=all`)] as const)
    );
    setSubAttributesByMain(Object.fromEntries(entries));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreateMain(e: FormEvent) {
    e.preventDefault();
    await api.post("/api/main-attributes", { key: newMainKey, name: newMainName });
    setNewMainKey("");
    setNewMainName("");
    await load();
  }

  async function handleCreateSub(mainAttributeId: number, e: FormEvent) {
    e.preventDefault();
    const name = newSubName[mainAttributeId];
    if (!name) return;
    await api.post("/api/sub-attributes", { main_attribute_id: mainAttributeId, name });
    setNewSubName((prev) => ({ ...prev, [mainAttributeId]: "" }));
    await load();
  }

  async function toggleSubActive(sub: SubAttribute) {
    if (sub.is_active) {
      await api.delete(`/api/sub-attributes/${sub.id}`);
      await load();
    }
  }

  return (
    <Container maxWidth="md" className="py-12">
      <Text as="h1" font="marker" size="2xl" weight="bold" className="mb-6">
        Attributes
      </Text>

      <Card variant="outlined" className="mb-8">
        <form onSubmit={handleCreateMain} className="flex flex-wrap gap-3">
          <Input
            className="flex-1"
            placeholder="key (e.g. delivery_speed)"
            value={newMainKey}
            onChange={(e) => setNewMainKey(e.target.value)}
            required
          />
          <Input
            className="flex-1"
            placeholder="Name (e.g. Delivery Speed)"
            value={newMainName}
            onChange={(e) => setNewMainName(e.target.value)}
            required
          />
          <Button type="submit">Add main attribute</Button>
        </form>
      </Card>

      <div className="flex flex-col gap-6">
        {mainAttributes.map((main) => (
          <Card key={main.id} variant="outlined">
            <Text as="h2" font="marker" size="lg" weight="semibold" className="mb-3">
              {main.name}
            </Text>
            <div className="mb-3 flex flex-col">
              {(subAttributesByMain[main.id] ?? []).map((sub, i) => (
                <div
                  key={sub.id}
                  className={`flex items-center justify-between py-2 ${i > 0 ? "border-t border-line" : ""}`}
                >
                  {sub.is_active ? (
                    <Text size="sm">{sub.name}</Text>
                  ) : (
                    <div className="flex items-center gap-2">
                      <Text size="sm" color="muted" className="line-through">
                        {sub.name}
                      </Text>
                      <Badge variant="subtle" color="default" size="sm">
                        inactive
                      </Badge>
                    </div>
                  )}
                  {sub.is_active && (
                    <Button variant="ghost" size="sm" onClick={() => toggleSubActive(sub)}>
                      Deactivate
                    </Button>
                  )}
                </div>
              ))}
            </div>
            <form onSubmit={(e) => handleCreateSub(main.id, e)} className="flex gap-2">
              <Input
                className="flex-1"
                size="sm"
                placeholder="New sub-attribute name"
                value={newSubName[main.id] ?? ""}
                onChange={(e) => setNewSubName((prev) => ({ ...prev, [main.id]: e.target.value }))}
              />
              <Button type="submit" size="sm">
                Add
              </Button>
            </form>
          </Card>
        ))}
      </div>
    </Container>
  );
}
