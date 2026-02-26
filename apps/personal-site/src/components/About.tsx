"use client";

import { Text, Container, Stack } from "@movoz/ui-web";

export function About() {
  return (
    <section id="about" className="py-32 px-6">
      <Container>
        {/* Content Grid */}
        <div className="grid lg:grid-cols-[1fr,1.1fr] gap-12 lg:gap-20 items-start">
          {/* Left - Image */}
          <div className="relative order-2 lg:order-1">
            <div className="aspect-[3/4] max-h-[500px] rounded-2xl overflow-hidden bg-zen-subtle shadow-lg">
              <img
                src="/images/profile.jpg"
                alt="To Ngoc Long"
                className="w-full h-full object-cover"
                onError={(e) => {
                  e.currentTarget.style.display = "none";
                  e.currentTarget.parentElement!.innerHTML =
                    '<div class="w-full h-full flex items-center justify-center text-zen-muted"><span class="text-lg">Your photo here</span></div>';
                }}
              />
            </div>
          </div>

          {/* Right - Header + Bio */}
          <div className="order-1 lg:order-2">
            {/* Header */}
            <div className="mb-8">
              <Text
                as="h2"
                font="serif"
                weight="bold"
                className="text-[2.5rem] md:text-[3.5rem] lg:text-[4rem] mb-4 leading-[1.05] tracking-[-0.02em]"
              >
                About me
              </Text>
              <Text className="md:text-[1.1rem] leading-[1.6] max-w-lg">
                A software engineer passionate about building products
                that make a difference and creating elegant solutions.
              </Text>
            </div>

            {/* Bio */}
            <Stack gap={5} className="max-w-lg">
              <Text className="md:text-[1.05rem] leading-[1.7]">
                Hello! I&apos;m a software engineer passionate about building
                products that make a difference. My journey in tech started
                with curiosity and has evolved into a career focused on
                creating elegant, efficient solutions.
              </Text>
              <Text className="md:text-[1.05rem] leading-[1.7]">
                I specialize in full-stack development, with a particular
                interest in developer experience, system design, and creating
                tools that help people work more effectively.
              </Text>
              <Text className="md:text-[1.05rem] leading-[1.7]">
                When I&apos;m not coding, you&apos;ll find me exploring new
                technologies, contributing to open source, or sharing knowledge
                with the developer community.
              </Text>
            </Stack>
          </div>
        </div>
      </Container>
    </section>
  );
}
