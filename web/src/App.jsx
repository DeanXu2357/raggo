import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Checkbox } from '@/components/ui/checkbox';

const KnowledgeBaseInterface = () => {
    const [selectedKB, setSelectedKB] = useState(null);
    const [selectedResources, setSelectedResources] = useState(new Set());

    // Mock data for demonstration
    const knowledgeBases = [
        {
            id: 1,
            name: 'Technical Documentation',
            resources: [
                { id: 1, name: 'API Guide.pdf', size: '1.2 MB' },
                { id: 2, name: 'Installation Manual.pdf', size: '850 KB' },
                { id: 3, name: 'Troubleshooting.md', size: '250 KB' }
            ]
        },
        {
            id: 2,
            name: 'Research Papers',
            resources: [
                { id: 4, name: 'Paper2024.pdf', size: '2.1 MB' },
                { id: 5, name: 'Research Notes.md', size: '450 KB' }
            ]
        }
    ];

    const handleKBSelect = (kbId) => {
        setSelectedKB(knowledgeBases.find(kb => kb.id.toString() === kbId));
        setSelectedResources(new Set());
    };

    const toggleResource = (resourceId) => {
        const newSelection = new Set(selectedResources);
        if (newSelection.has(resourceId)) {
            newSelection.delete(resourceId);
        } else {
            newSelection.add(resourceId);
        }
        setSelectedResources(newSelection);
    };

    const toggleAllResources = () => {
        if (!selectedKB) return;

        if (selectedResources.size === selectedKB.resources.length) {
            setSelectedResources(new Set());
        } else {
            setSelectedResources(new Set(selectedKB.resources.map(r => r.id)));
        }
    };

    return (
        <div className="min-h-screen p-4 bg-slate-950">
            <div className="max-w-6xl mx-auto">
                <div className="flex gap-4">
                    {/* Chat Interface */}
                    <Card className="flex-1 h-[600px] flex flex-col bg-slate-900 border-slate-800">
                        <CardHeader className="border-b border-slate-800">
                            <div className="flex items-center justify-between">
                                <CardTitle className="text-slate-200 text-xl">Chat Interface</CardTitle>
                                <Select onValueChange={handleKBSelect} value={selectedKB?.id.toString()}>
                                    <SelectTrigger className="w-64 bg-slate-800 border-slate-700 text-slate-200">
                                        <SelectValue placeholder="Select Knowledge Base" />
                                    </SelectTrigger>
                                    <SelectContent className="bg-slate-800 border-slate-700">
                                        {knowledgeBases.map(kb => (
                                            <SelectItem
                                                key={kb.id}
                                                value={kb.id.toString()}
                                                className="text-slate-200 hover:bg-slate-700 focus:bg-slate-700"
                                            >
                                                {kb.name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                        </CardHeader>
                        <CardContent className="flex-1 overflow-y-auto p-4">
                            <div className="text-center text-slate-400">
                                {selectedKB
                                    ? `Connected to ${selectedKB.name}`
                                    : 'Please select a knowledge base to start chatting'
                                }
                            </div>
                        </CardContent>
                        <div className="p-4 border-t border-slate-800">
                            <div className="flex space-x-2">
                                <Input
                                    placeholder={selectedKB ? "Type your message..." : "Select a knowledge base first"}
                                    className="flex-1 bg-slate-800 border-slate-700 text-slate-200 placeholder:text-slate-500"
                                    disabled={!selectedKB}
                                />
                                <Button
                                    variant="default"
                                    className="bg-blue-600 hover:bg-blue-500 text-slate-100"
                                    disabled={!selectedKB}
                                >
                                    Send
                                </Button>
                            </div>
                        </div>
                    </Card>

                    {/* Resources Panel */}
                    <Card className="w-72 h-[600px] bg-slate-900 border-slate-800">
                        <CardHeader className="border-b border-slate-800">
                            <div className="flex items-center justify-between">
                                <CardTitle className="text-lg text-slate-200">Available Resources</CardTitle>
                                {selectedKB && (
                                    <div className="border border-slate-700 rounded p-0.5">
                                        <Checkbox
                                            checked={selectedKB && selectedResources.size === selectedKB.resources.length}
                                            onCheckedChange={toggleAllResources}
                                            className="bg-slate-800 border-slate-700 data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
                                        />
                                    </div>
                                )}
                            </div>
                        </CardHeader>
                        <ScrollArea className="h-[520px]">
                            <CardContent className="p-4">
                                {selectedKB ? (
                                    <div className="space-y-2">
                                        {selectedKB.resources.map(resource => (
                                            <div
                                                key={resource.id}
                                                className="p-2 rounded-lg bg-slate-800 border border-slate-700 flex items-center gap-2 hover:bg-slate-750"
                                            >
                                                <div className="border border-slate-700 rounded p-0.5">
                                                    <Checkbox
                                                        checked={selectedResources.has(resource.id)}
                                                        onCheckedChange={() => toggleResource(resource.id)}
                                                        className="bg-slate-800 border-slate-700 data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
                                                    />
                                                </div>
                                                <div className="flex-1 min-w-0">
                                                    <div className="font-medium text-slate-200 truncate">
                                                        {resource.name}
                                                    </div>
                                                    <div className="text-sm text-slate-400">
                                                        {resource.size}
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-center text-slate-400">
                                        Select a knowledge base to view resources
                                    </div>
                                )}
                            </CardContent>
                        </ScrollArea>
                    </Card>
                </div>
            </div>
        </div>
    );
};

export default KnowledgeBaseInterface;
