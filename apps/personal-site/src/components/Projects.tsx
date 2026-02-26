"use client";

import { useState } from "react";
import { Text, Container, Stack } from "@movoz/ui-web";

interface Project {
  id: string;
  title: string;
  description: string;
  media?: string;
  mediaType?: "video" | "image";
}

const projects: Project[] = [
  {
    id: "project-1",
    title: "Workspaces",
    description:
      "Organize your tabs into Workspaces to keep your projects separate and organized, and switch between them with ease.",
    media: "/projects/workspace.mp4",
    mediaType: "video",
  },
  {
    id: "project-2",
    title: "Compact Mode",
    description:
      "Zen's Compact Mode gives you more screen real estate by hiding the tab bar when you don't need it, and showing it when you do.",
    media: "/projects/compact.mp4",
    mediaType: "video",
  },
  {
    id: "project-3",
    title: "Glance",
    description:
      "Glance allows you to quickly switch between your most used tabs, without having to scroll through your history.",
    media: "/projects/glance.mp4",
    mediaType: "video",
  },
  {
    id: "project-4",
    title: "Split View",
    description:
      "Split View allows you to view two tabs side by side, making it easier to compare and switch between them.",
    media: "/projects/split.mp4",
    mediaType: "video",
  },
];

export function Projects() {
  const [activeProject, setActiveProject] = useState(projects[0]);

  return (
    <section id="projects" className="py-32 px-6 overflow-hidden">
      <Container>
        {/* Header */}
        <div className="mb-10 max-w-2xl">
          <Text
            as="h2"
            font="serif"
            weight="bold"
            className="text-[2rem] md:text-[3rem] lg:text-[4rem] mb-4 leading-[1.05] tracking-[-0.02em] whitespace-nowrap"
          >
            Introducing the citizens
          </Text>
          <Text className="md:text-[1.1rem] leading-[1.6]">
            Zen is packed with features that help you stay productive
            <br className="hidden md:block" />
            and focused. Browsers should be tools that help you get
            <br className="hidden md:block" />
            things done, not distractions that keep you from your work.
          </Text>
        </div>

        {/* Content Grid */}
        <div className="grid lg:grid-cols-[1fr,1.2fr] gap-8 lg:gap-6 items-start">
          {/* Left - Project List */}
          <Stack gap={4}>
            {projects.map((project) => (
              <button
                key={project.id}
                onClick={() => setActiveProject(project)}
                className={`w-full text-left px-5 py-4 rounded-lg transition-all duration-200 border-l-[3px] ${
                  activeProject.id === project.id
                    ? "bg-zen-subtle border-l-zen-text"
                    : "border-l-transparent hover:bg-zen-subtle/50"
                }`}
              >
                <Text
                  as="h3"
                  weight="bold"
                  className="text-[1.4rem] md:text-[1.6rem] mb-1 leading-tight"
                >
                  {project.title}
                </Text>
                <Text className="text-[0.95rem] md:text-base leading-[1.5]">
                  {project.description}
                </Text>
              </button>
            ))}
          </Stack>

          {/* Right - Media Display */}
          <div className="relative lg:ml-4">
            <div className="aspect-[4/3] rounded-2xl overflow-hidden bg-[#8b7b7b] shadow-2xl lg:translate-x-8 xl:translate-x-16">
              {activeProject.media ? (
                activeProject.mediaType === "video" ? (
                  <video
                    key={activeProject.id}
                    src={activeProject.media}
                    autoPlay
                    loop
                    muted
                    playsInline
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <img
                    src={activeProject.media}
                    alt={activeProject.title}
                    className="w-full h-full object-cover"
                  />
                )
              ) : (
                <div className="w-full h-full flex items-center justify-center text-white/60">
                  <Text as="span" size="lg">
                    Video/Image placeholder
                  </Text>
                </div>
              )}
            </div>
          </div>
        </div>
      </Container>
    </section>
  );
}
