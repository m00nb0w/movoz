"use client";

import { Mail, MapPin, Send, Github, Linkedin, Twitter } from "lucide-react";
import {
  Button,
  Input,
  IconButton,
  Text,
  Stack,
  Container,
} from "@movoz/ui-web";

const socialLinks = [
  { icon: Github, href: "https://github.com", label: "GitHub" },
  { icon: Linkedin, href: "https://linkedin.com", label: "LinkedIn" },
  { icon: Twitter, href: "https://twitter.com", label: "Twitter" },
];

export function Contact() {
  return (
    <section id="contact" className="py-32 px-6">
      <Container>
        {/* Header */}
        <div className="mb-12 max-w-2xl">
          <Text
            as="h2"
            font="serif"
            weight="bold"
            className="text-[2.5rem] md:text-[3.5rem] lg:text-[4rem] leading-[1.05] tracking-[-0.02em] mb-4"
          >
            Get in touch
          </Text>
          <Text size="base" className="md:text-[1.1rem] leading-[1.6]">
            Have a project in mind or just want to chat?
            <br className="hidden md:block" />
            I&apos;d love to hear from you.
          </Text>
        </div>

        {/* Content Grid */}
        <div className="grid lg:grid-cols-2 gap-12 lg:gap-16">
          {/* Left - Contact Info */}
          <Stack gap={8}>
            <div>
              <Text as="h3" weight="bold" className="text-[1.4rem] md:text-[1.6rem] mb-4 leading-tight">
                Contact
              </Text>
              <Stack gap={4}>
                <div className="flex items-center gap-4">
                  <Mail className="w-5 h-5 text-zen-text" strokeWidth={1.5} />
                  <a
                    href="mailto:hello@example.com"
                    className="text-zen-text text-base hover:text-accent transition-colors"
                  >
                    hello@example.com
                  </a>
                </div>
                <div className="flex items-center gap-4">
                  <MapPin className="w-5 h-5 text-zen-text" strokeWidth={1.5} />
                  <Text as="span">San Francisco, CA</Text>
                </div>
              </Stack>
            </div>

            <div>
              <Text as="h3" weight="bold" className="text-[1.4rem] md:text-[1.6rem] mb-4 leading-tight">
                Connect
              </Text>
              <Stack direction="horizontal" gap={4}>
                {socialLinks.map((link) => (
                  <a
                    key={link.label}
                    href={link.href}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <IconButton
                      icon={<link.icon className="w-5 h-5" />}
                      variant="secondary"
                      label={link.label}
                    />
                  </a>
                ))}
              </Stack>
            </div>
          </Stack>

          {/* Right - Contact Form */}
          <div>
            <Text as="h3" weight="bold" className="text-[1.4rem] md:text-[1.6rem] mb-6 leading-tight">
              Send a message
            </Text>
            <form>
              <Stack gap={5}>
                <Input label="Name" name="name" placeholder="Your name" />
                <Input
                  label="Email"
                  name="email"
                  type="email"
                  placeholder="your@email.com"
                />
                <div className="w-full">
                  <label
                    htmlFor="message"
                    className="block font-semibold text-zen-text text-base mb-2"
                  >
                    Message
                  </label>
                  <textarea
                    id="message"
                    name="message"
                    rows={5}
                    className="w-full px-4 py-3 bg-zen-subtle border border-zen-border rounded-lg focus:outline-none focus:ring-2 focus:ring-zen-text/20 focus:border-zen-text text-zen-text placeholder-zen-muted resize-none text-base transition-all"
                    placeholder="Your message..."
                  />
                </div>
                <Button
                  type="submit"
                  iconRight={<Send className="w-4 h-4" />}
                >
                  Send Message
                </Button>
              </Stack>
            </form>
          </div>
        </div>
      </Container>
    </section>
  );
}
