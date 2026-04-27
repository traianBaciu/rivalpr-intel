"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
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
import { Plus, Download } from "lucide-react";

export default function JournalistsPage() {
  // === Outlets state ===
  const [outlets, setOutlets] = useState<Outlet[]>([]);
  const [outletForm, setOutletForm] = useState<CreateOutletRequest>({
    name: "",
    website: "",
    country: "",
  });
  const [outletDialogOpen, setOutletDialogOpen] = useState(false);

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
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  // === Import dialog state ===
  const [importDialogOpen, setImportDialogOpen] = useState(false);
  const [importSearch, setImportSearch] = useState("");
  const [importNiche, setImportNiche] = useState("all");
  const [agencyJournalists, setAgencyJournalists] = useState<Journalist[]>([]);

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

  const fetchAgencyJournalists = useCallback(async () => {
    try {
      const res = await api.get<PaginatedResponse<Journalist>>(
        `/api/journalists/agency?limit=200`
      );
      setAgencyJournalists(res?.data || []);
    } catch {
      toast.error("Failed to load agency journalists");
    }
  }, []);

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

  // Fetch agency pool when import dialog opens
  useEffect(() => {
    if (importDialogOpen) fetchAgencyJournalists();
  }, [importDialogOpen, fetchAgencyJournalists]);

  // === Outlet handlers ===
  function openCreateOutlet() {
    setOutletForm({ name: "", website: "", country: "" });
    setOutletDialogOpen(true);
  }

  async function handleOutletSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await api.post("/api/outlets", outletForm);
      toast.success("Outlet created");
      setOutletDialogOpen(false);
      fetchOutlets();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Operation failed");
    }
  }

  // === Journalist handlers ===
  function openCreateJournalist() {
    setJournalistForm({ outlet_id: "", name: "", email: "", niche: "" });
    setJournalistDialogOpen(true);
  }

  async function handleJournalistSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await api.post("/api/journalists", journalistForm);
      toast.success("Journalist added to agency database");
      setJournalistDialogOpen(false);
      setPage(1);
      fetchJournalists(1);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Operation failed");
    }
  }

  // === CRM handlers ===
  function openCreateCrm(journalistId: string) {
    setCrmForm({
      journalist_id: journalistId,
      relationship_score: 5,
      private_notes: "",
    });
    setImportDialogOpen(false);
    setCrmDialogOpen(true);
  }

  async function handleImport(journalistId: string) {
    try {
      await api.post("/api/crm/relationships", { journalist_id: journalistId, relationship_score: 0 });
      toast.success("Journalist added to your list");
      await Promise.all([fetchJournalists(1), fetchAgencyJournalists(), fetchRelationships()]);
      setPage(1);
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Import failed");
    }
  }

  async function handleCrmSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      const existing = relationships.find((r) => r.journalist_id === crmForm.journalist_id);
      if (existing) {
        await api.put(`/api/crm/relationships/${existing.id}`, {
          relationship_score: crmForm.relationship_score,
          private_notes: crmForm.private_notes,
        });
      } else {
        await api.post("/api/crm/relationships", crmForm);
      }
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

  // === Import dialog computed values ===
  const allNiches = useMemo(() => {
    const niches = new Set(agencyJournalists.map((j) => j.niche).filter(Boolean));
    return Array.from(niches).sort();
  }, [agencyJournalists]);

  const importList = useMemo(() => {
    return agencyJournalists.filter((j) => {
      if (importNiche !== "all" && j.niche !== importNiche) return false;
      if (importSearch) {
        const q = importSearch.toLowerCase();
        return (
          j.name.toLowerCase().includes(q) ||
          j.email.toLowerCase().includes(q)
        );
      }
      return true;
    });
  }, [agencyJournalists, importNiche, importSearch]);

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
          <div className="flex justify-end gap-2">
            {/* Import from Agency */}
            <Dialog open={importDialogOpen} onOpenChange={setImportDialogOpen}>
              <DialogTrigger render={<Button variant="outline" />}>
                <Download className="mr-2 h-4 w-4" /> Import from Agency
              </DialogTrigger>
              <DialogContent className="max-w-2xl">
                <DialogHeader>
                  <DialogTitle>Import from Agency Database</DialogTitle>
                </DialogHeader>
                <div className="space-y-3">
                  <div className="flex gap-2">
                    <Input
                      placeholder="Search by name or email…"
                      value={importSearch}
                      onChange={(e) => setImportSearch(e.target.value)}
                      className="flex-1"
                    />
                    <Select
                      value={importNiche}
                      onValueChange={(v) => setImportNiche(v ?? "all")}
                    >
                      <SelectTrigger className="w-48">
                        <SelectValue placeholder="All niches" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All niches</SelectItem>
                        {allNiches.map((n) => (
                          <SelectItem key={n} value={n}>
                            {n}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="max-h-96 overflow-y-auto rounded-md border">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Name</TableHead>
                          <TableHead>Email</TableHead>
                          <TableHead>Outlet</TableHead>
                          <TableHead>Niche</TableHead>
                          <TableHead className="w-20"></TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {importList.length === 0 ? (
                          <TableRow>
                            <TableCell
                              colSpan={5}
                              className="text-center text-muted-foreground py-8"
                            >
                              {agencyJournalists.length === 0
                                ? "You've already tracked all journalists in the agency database."
                                : "No journalists match your search."}
                            </TableCell>
                          </TableRow>
                        ) : (
                          importList.map((j) => (
                            <TableRow key={j.id}>
                              <TableCell className="font-medium">{j.name}</TableCell>
                              <TableCell className="text-sm text-muted-foreground">
                                {j.email}
                              </TableCell>
                              <TableCell>{j.outlet?.name || "—"}</TableCell>
                              <TableCell>{j.niche || "—"}</TableCell>
                              <TableCell>
                                <Button size="sm" onClick={() => handleImport(j.id)}>
                                  Import
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))
                        )}
                      </TableBody>
                    </Table>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Showing {importList.length} untracked journalist
                    {importList.length !== 1 ? "s" : ""}.
                  </p>                </div>
              </DialogContent>
            </Dialog>

            {/* Add Journalist */}
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
                  <DialogTitle>Add Journalist to Agency Database</DialogTitle>
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
                    Add to Agency Database
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
                          {rel && rel.relationship_score > 0 ? (
                            <div className="flex items-center gap-2 min-w-35">
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

          {/* CRM Rate Dialog */}
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
                    Add Outlet to Agency Database
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
                    Add to Agency Database
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
