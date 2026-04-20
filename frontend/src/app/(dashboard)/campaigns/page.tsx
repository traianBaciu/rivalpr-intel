"use client";

import { useCallback, useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import type { Campaign, Client, CreateCampaignRequest } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
import { toast } from "sonner";
import { Plus, Pencil, Trash2 } from "lucide-react";

export default function CampaignsPage() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingCampaign, setEditingCampaign] = useState<Campaign | null>(
    null
  );
  const [form, setForm] = useState<CreateCampaignRequest>({
    client_id: "",
    title: "",
    press_release_text: "",
  });

  const fetchCampaigns = useCallback(async () => {
    try {
      const data = await api.get<Campaign[]>("/api/campaigns");
      setCampaigns(data || []);
    } catch {
      toast.error("Failed to load campaigns");
    }
  }, []);

  const fetchClients = useCallback(async () => {
    try {
      const data = await api.get<Client[]>("/api/clients");
      setClients(data || []);
    } catch {
      // handled
    }
  }, []);

  useEffect(() => {
    Promise.all([fetchCampaigns(), fetchClients()]).finally(() =>
      setLoading(false)
    );
  }, [fetchCampaigns, fetchClients]);

  function openCreate() {
    setEditingCampaign(null);
    setForm({ client_id: "", title: "", press_release_text: "" });
    setDialogOpen(true);
  }

  function openEdit(campaign: Campaign) {
    setEditingCampaign(campaign);
    setForm({
      client_id: campaign.client_id,
      title: campaign.title,
      press_release_text: campaign.press_release_text,
    });
    setDialogOpen(true);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      if (editingCampaign) {
        await api.put(`/api/campaigns/${editingCampaign.id}`, {
          title: form.title,
          press_release_text: form.press_release_text,
        });
        toast.success("Campaign updated");
      } else {
        await api.post("/api/campaigns", form);
        toast.success("Campaign created");
      }
      setDialogOpen(false);
      fetchCampaigns();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Operation failed");
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this campaign?")) return;
    try {
      await api.del(`/api/campaigns/${id}`);
      toast.success("Campaign deleted");
      fetchCampaigns();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Delete failed");
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Campaigns</h1>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger
            render={<Button />}
            onClick={openCreate}
          >
            <Plus className="mr-2 h-4 w-4" /> New Campaign
          </DialogTrigger>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>
                {editingCampaign ? "Edit Campaign" : "New Campaign"}
              </DialogTitle>
            </DialogHeader>
            <form onSubmit={handleSubmit} className="space-y-4">
              {!editingCampaign && (
                <div className="space-y-2">
                  <Label>Client</Label>
                  <Select
                    value={form.client_id}
                    onValueChange={(v) =>
                      setForm({ ...form, client_id: v ?? "" })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select client" />
                    </SelectTrigger>
                    <SelectContent>
                      {clients.map((c) => (
                        <SelectItem key={c.id} value={c.id}>
                          {c.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
              <div className="space-y-2">
                <Label>Title</Label>
                <Input
                  value={form.title}
                  onChange={(e) =>
                    setForm({ ...form, title: e.target.value })
                  }
                  required
                />
              </div>
              <div className="space-y-2">
                <Label>Press Release Text</Label>
                <Textarea
                  value={form.press_release_text || ""}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      press_release_text: e.target.value,
                    })
                  }
                  rows={8}
                  placeholder="Paste the full press release here..."
                />
              </div>
              <Button type="submit" className="w-full">
                {editingCampaign ? "Update" : "Create"}
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Campaigns</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p className="text-muted-foreground">Loading...</p>
          ) : campaigns.length === 0 ? (
            <p className="text-muted-foreground">
              No campaigns yet. Create one above.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Title</TableHead>
                  <TableHead>Client</TableHead>
                  <TableHead>Press Release</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="w-24">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {campaigns.map((campaign) => (
                  <TableRow key={campaign.id}>
                    <TableCell className="font-medium">
                      {campaign.title}
                    </TableCell>
                    <TableCell>{campaign.client?.name || "—"}</TableCell>
                    <TableCell className="max-w-xs truncate">
                      {campaign.press_release_text
                        ? campaign.press_release_text.substring(0, 80) + "..."
                        : "—"}
                    </TableCell>
                    <TableCell>
                      {new Date(campaign.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openEdit(campaign)}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(campaign.id)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
