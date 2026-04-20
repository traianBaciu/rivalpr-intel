"use client";

import { useCallback, useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import type {
  Journalist,
  Outlet,
  CrmRelationship,
  CreateOutletRequest,
  CreateJournalistRequest,
  CreateCrmRelationshipRequest,
  PaginatedResponse,
} from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Slider } from "@/components/ui/slider";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "sonner";
import { Plus, Pencil, Trash2 } from "lucide-react";

export default function JournalistsPage() {
  // === Outlets state ===
  const [outlets, setOutlets] = useState<Outlet[]>([]);
  const [outletForm, setOutletForm] = useState<CreateOutletRequest>({
    name: "",
    website: "",
    country: "",
  });
  const [outletDialogOpen, setOutletDialogOpen] = useState(false);
  const [editingOutlet, setEditingOutlet] = useState<Outlet | null>(null);

  // === Journalists state ===
  const [journalists, setJournalists] = useState<Journalist[]>([]);
  const [journalistForm, setJournalistForm] =
    useState<CreateJournalistRequest>({
      outlet_id: "",
      name: "",
      email: "",
      niche: "",
    });
  const [journalistDialogOpen, setJournalistDialogOpen] = useState(false);
  const [editingJournalist, setEditingJournalist] =
    useState<Journalist | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  // === CRM state ===
  const [relationships, setRelationships] = useState<CrmRelationship[]>([]);
  const [crmDialogOpen, setCrmDialogOpen] = useState(false);
  const [crmForm, setCrmForm] = useState<CreateCrmRelationshipRequest>({
    journalist_id: "",
    relationship_score: 5,
    private_notes: "",
  });

  const [loading, setLoading] = useState(true);

  const fetchOutlets = useCallback(async () => {
    try {
      const data = await api.get<Outlet[]>("/api/outlets");
      setOutlets(data || []);
    } catch {
      toast.error("Failed to load outlets");
    }
  }, []);

  const fetchJournalists = useCallback(
    async (p = 1) => {
      try {
        const res = await api.get<PaginatedResponse<Journalist>>(
          `/api/journalists?page=${p}&limit=20`
        );
        const list = res?.data || [];
        if (p === 1) {
          setJournalists(list);
        } else {
          setJournalists((prev) => [...prev, ...list]);
        }
        setHasMore(list.length === 20);
      } catch {
        toast.error("Failed to load journalists");
      }
    },
    []
  );

  const fetchRelationships = useCallback(async () => {
    try {
      const data = await api.get<CrmRelationship[]>("/api/crm/relationships");
      setRelationships(data || []);
    } catch {
      toast.error("Failed to load relationships");
    }
  }, []);

  useEffect(() => {
    Promise.all([fetchOutlets(), fetchJournalists(1), fetchRelationships()]).finally(() =>
      setLoading(false)
    );
  }, [fetchOutlets, fetchJournalists, fetchRelationships]);

  // === Outlet handlers ===
  function openCreateOutlet() {
    setEditingOutlet(null);
    setOutletForm({ name: "", website: "", country: "" });
    setOutletDialogOpen(true);
  }

  function openEditOutlet(outlet: Outlet) {
    setEditingOutlet(outlet);
    setOutletForm({
      name: outlet.name,
      website: outlet.website,
      country: outlet.country,
    });
    setOutletDialogOpen(true);
  }

  async function handleOutletSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      if (editingOutlet) {
        await api.put(`/api/outlets/${editingOutlet.id}`, outletForm);
        toast.success("Outlet updated");
      } else {
        await api.post("/api/outlets", outletForm);
        toast.success("Outlet created");
      }
      setOutletDialogOpen(false);
      fetchOutlets();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Operation failed");
    }
  }

  async function handleDeleteOutlet(id: string) {
    if (!confirm("Delete this outlet?")) return;
    try {
      await api.del(`/api/outlets/${id}`);
      toast.success("Outlet deleted");
      fetchOutlets();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Delete failed");
    }
  }

  // === Journalist handlers ===
  function openCreateJournalist() {
    setEditingJournalist(null);
    setJournalistForm({ outlet_id: "", name: "", email: "", niche: "" });
    setJournalistDialogOpen(true);
  }

  function openEditJournalist(j: Journalist) {
    setEditingJournalist(j);
    setJournalistForm({
      outlet_id: j.outlet_id,
      name: j.name,
      email: j.email,
      niche: j.niche,
    });
    setJournalistDialogOpen(true);
  }

  async function handleJournalistSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      if (editingJournalist) {
        await api.put(
          `/api/journalists/${editingJournalist.id}`,
          journalistForm
        );
        toast.success("Journalist updated");
      } else {
        await api.post("/api/journalists", journalistForm);
        toast.success("Journalist created");
      }
      setJournalistDialogOpen(false);
      setPage(1);
      fetchJournalists(1);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Operation failed");
    }
  }

  async function handleDeleteJournalist(id: string) {
    if (!confirm("Delete this journalist?")) return;
    try {
      await api.del(`/api/journalists/${id}`);
      toast.success("Journalist deleted");
      setPage(1);
      fetchJournalists(1);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Delete failed");
    }
  }

  // === CRM handlers ===
  function openCreateCrm(journalistId: string) {
    setCrmForm({
      journalist_id: journalistId,
      relationship_score: 5,
      private_notes: "",
    });
    setCrmDialogOpen(true);
  }

  async function handleCrmSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await api.post("/api/crm/relationships", crmForm);
      toast.success("Relationship saved");
      setCrmDialogOpen(false);
      fetchRelationships();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Operation failed");
    }
  }

  async function handleCrmScoreUpdate(id: string, score: number) {
    try {
      await api.put(`/api/crm/relationships/${id}`, {
        relationship_score: score,
      });
      fetchRelationships();
    } catch {
      toast.error("Failed to update score");
    }
  }

  function getRelationship(journalistId: string) {
    return relationships.find((r) => r.journalist_id === journalistId);
  }

  if (loading) {
    return <div className="text-muted-foreground">Loading...</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Journalists & Outlets</h1>

      <Tabs defaultValue="journalists">
        <TabsList>
          <TabsTrigger value="journalists">Journalists</TabsTrigger>
          <TabsTrigger value="outlets">Outlets</TabsTrigger>
        </TabsList>

        {/* === JOURNALISTS TAB === */}
        <TabsContent value="journalists" className="space-y-4">
          <div className="flex justify-end">
            <Dialog
              open={journalistDialogOpen}
              onOpenChange={setJournalistDialogOpen}
            >
              <DialogTrigger
                render={<Button />}
                onClick={openCreateJournalist}
              >
                <Plus className="mr-2 h-4 w-4" /> Add Journalist
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>
                    {editingJournalist
                      ? "Edit Journalist"
                      : "Add Journalist"}
                  </DialogTitle>
                </DialogHeader>
                <form
                  onSubmit={handleJournalistSubmit}
                  className="space-y-4"
                >
                  <div className="space-y-2">
                    <Label>Outlet</Label>
                    <Select
                      value={journalistForm.outlet_id}
                      onValueChange={(v) =>
                        setJournalistForm({ ...journalistForm, outlet_id: v ?? "" })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select outlet" />
                      </SelectTrigger>
                      <SelectContent>
                        {outlets.map((o) => (
                          <SelectItem key={o.id} value={o.id}>
                            {o.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>Name</Label>
                    <Input
                      value={journalistForm.name}
                      onChange={(e) =>
                        setJournalistForm({
                          ...journalistForm,
                          name: e.target.value,
                        })
                      }
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Email</Label>
                    <Input
                      type="email"
                      value={journalistForm.email}
                      onChange={(e) =>
                        setJournalistForm({
                          ...journalistForm,
                          email: e.target.value,
                        })
                      }
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Niche / Beat</Label>
                    <Input
                      value={journalistForm.niche || ""}
                      onChange={(e) =>
                        setJournalistForm({
                          ...journalistForm,
                          niche: e.target.value,
                        })
                      }
                    />
                  </div>
                  <Button type="submit" className="w-full">
                    {editingJournalist ? "Update" : "Create"}
                  </Button>
                </form>
              </DialogContent>
            </Dialog>
          </div>

          <Card>
            <CardContent className="pt-6">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Outlet</TableHead>
                    <TableHead>Niche</TableHead>
                    <TableHead>CRM Score</TableHead>
                    <TableHead className="w-24">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {journalists.map((j) => {
                    const rel = getRelationship(j.id);
                    return (
                      <TableRow key={j.id}>
                        <TableCell className="font-medium">
                          {j.name}
                        </TableCell>
                        <TableCell>{j.email}</TableCell>
                        <TableCell>{j.outlet?.name || "—"}</TableCell>
                        <TableCell>{j.niche || "—"}</TableCell>
                        <TableCell>
                          {rel ? (
                            <div className="flex items-center gap-2 min-w-[140px]">
                              <Slider
                                value={[rel.relationship_score]}
                                min={1}
                                max={10}
                                step={1}
                                className="w-20"
                                onValueCommitted={(v) =>
                                  handleCrmScoreUpdate(rel.id, Array.isArray(v) ? v[0] : v)
                                }
                              />
                              <Badge variant="secondary">
                                {rel.relationship_score}
                              </Badge>
                            </div>
                          ) : (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => openCreateCrm(j.id)}
                            >
                              Rate
                            </Button>
                          )}
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => openEditJournalist(j)}
                            >
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDeleteJournalist(j.id)}
                            >
                              <Trash2 className="h-4 w-4 text-destructive" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
              {hasMore && (
                <div className="mt-4 text-center">
                  <Button
                    variant="outline"
                    onClick={() => {
                      const next = page + 1;
                      setPage(next);
                      fetchJournalists(next);
                    }}
                  >
                    Load More
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          {/* CRM Dialog */}
          <Dialog open={crmDialogOpen} onOpenChange={setCrmDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Rate Relationship</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleCrmSubmit} className="space-y-4">
                <div className="space-y-2">
                  <Label>
                    Score: {crmForm.relationship_score}
                  </Label>
                  <Slider
                    value={[crmForm.relationship_score]}
                    min={1}
                    max={10}
                    step={1}
                    onValueChange={(v) =>
                      setCrmForm({
                        ...crmForm,
                        relationship_score: Array.isArray(v) ? v[0] : v,
                      })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Private Notes</Label>
                  <Textarea
                    value={crmForm.private_notes || ""}
                    onChange={(e) =>
                      setCrmForm({
                        ...crmForm,
                        private_notes: e.target.value,
                      })
                    }
                    rows={3}
                  />
                </div>
                <Button type="submit" className="w-full">
                  Save
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        </TabsContent>

        {/* === OUTLETS TAB === */}
        <TabsContent value="outlets" className="space-y-4">
          <div className="flex justify-end">
            <Dialog
              open={outletDialogOpen}
              onOpenChange={setOutletDialogOpen}
            >
              <DialogTrigger
                render={<Button />}
                onClick={openCreateOutlet}
              >
                <Plus className="mr-2 h-4 w-4" /> Add Outlet
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>
                    {editingOutlet ? "Edit Outlet" : "Add Outlet"}
                  </DialogTitle>
                </DialogHeader>
                <form onSubmit={handleOutletSubmit} className="space-y-4">
                  <div className="space-y-2">
                    <Label>Name</Label>
                    <Input
                      value={outletForm.name}
                      onChange={(e) =>
                        setOutletForm({
                          ...outletForm,
                          name: e.target.value,
                        })
                      }
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Website</Label>
                    <Input
                      value={outletForm.website || ""}
                      onChange={(e) =>
                        setOutletForm({
                          ...outletForm,
                          website: e.target.value,
                        })
                      }
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Country</Label>
                    <Input
                      value={outletForm.country || ""}
                      onChange={(e) =>
                        setOutletForm({
                          ...outletForm,
                          country: e.target.value,
                        })
                      }
                    />
                  </div>
                  <Button type="submit" className="w-full">
                    {editingOutlet ? "Update" : "Create"}
                  </Button>
                </form>
              </DialogContent>
            </Dialog>
          </div>

          <Card>
            <CardContent className="pt-6">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Website</TableHead>
                    <TableHead>Country</TableHead>
                    <TableHead className="w-24">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {outlets.map((outlet) => (
                    <TableRow key={outlet.id}>
                      <TableCell className="font-medium">
                        {outlet.name}
                      </TableCell>
                      <TableCell>
                        {outlet.website ? (
                          <a
                            href={outlet.website}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-primary underline"
                          >
                            {outlet.website}
                          </a>
                        ) : (
                          "—"
                        )}
                      </TableCell>
                      <TableCell>{outlet.country || "—"}</TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openEditOutlet(outlet)}
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleDeleteOutlet(outlet.id)}
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
