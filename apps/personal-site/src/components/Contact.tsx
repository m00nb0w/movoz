"use client";

import { Mail, MapPin, Send, Github, Linkedin, Twitter } from "lucide-react";

const socialLinks = [
  { icon: Github, href: "https://github.com", label: "GitHub" },
  { icon: Linkedin, href: "https://linkedin.com", label: "LinkedIn" },
  { icon: Twitter, href: "https://twitter.com", label: "Twitter" },
];

export function Contact() {
  return (
    <section id="contact" className="py-32 px-6">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-12 max-w-2xl">
          <h2 className="font-serif text-[2.5rem] md:text-[3.5rem] lg:text-[4rem] font-bold text-zen-text mb-4 leading-[1.05] tracking-[-0.02em]">
            Get in touch
          </h2>
          <p className="text-zen-text text-base md:text-[1.1rem] leading-[1.6] font-normal">
            Have a project in mind or just want to chat?
            <br className="hidden md:block" />
            I&apos;d love to hear from you.
          </p>
        </div>

        {/* Content Grid */}
        <div className="grid lg:grid-cols-2 gap-12 lg:gap-16">
          {/* Left - Contact Info */}
          <div className="space-y-8">
            <div>
              <h3 className="font-bold text-[1.4rem] md:text-[1.6rem] text-zen-text mb-4 leading-tight">
                Contact
              </h3>
              <div className="space-y-4">
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
                  <span className="text-zen-text text-base">San Francisco, CA</span>
                </div>
              </div>
            </div>

            <div>
              <h3 className="font-bold text-[1.4rem] md:text-[1.6rem] text-zen-text mb-4 leading-tight">
                Connect
              </h3>
              <div className="flex gap-4">
                {socialLinks.map((link) => (
                  <a
                    key={link.label}
                    href={link.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="p-3 bg-zen-subtle rounded-lg hover:bg-zen-text hover:text-zen-bg transition-all duration-200"
                    aria-label={link.label}
                  >
                    <link.icon className="w-5 h-5" />
                  </a>
                ))}
              </div>
            </div>
          </div>

          {/* Right - Contact Form */}
          <div>
            <h3 className="font-bold text-[1.4rem] md:text-[1.6rem] text-zen-text mb-6 leading-tight">
              Send a message
            </h3>
            <form className="space-y-5">
              <div>
                <label
                  htmlFor="name"
                  className="block font-semibold text-zen-text text-base mb-2"
                >
                  Name
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  className="w-full px-4 py-3 bg-zen-subtle border border-zen-border rounded-lg focus:outline-none focus:ring-2 focus:ring-zen-text/20 focus:border-zen-text text-zen-text placeholder-zen-muted text-base transition-all"
                  placeholder="Your name"
                />
              </div>

              <div>
                <label
                  htmlFor="email"
                  className="block font-semibold text-zen-text text-base mb-2"
                >
                  Email
                </label>
                <input
                  type="email"
                  id="email"
                  name="email"
                  className="w-full px-4 py-3 bg-zen-subtle border border-zen-border rounded-lg focus:outline-none focus:ring-2 focus:ring-zen-text/20 focus:border-zen-text text-zen-text placeholder-zen-muted text-base transition-all"
                  placeholder="your@email.com"
                />
              </div>

              <div>
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

              <button
                type="submit"
                className="inline-flex items-center gap-2 px-6 py-3 bg-zen-text text-zen-bg text-base font-medium rounded-lg hover:opacity-90 transition-opacity duration-200"
              >
                Send Message
                <Send className="w-4 h-4" />
              </button>
            </form>
          </div>
        </div>
      </div>
    </section>
  );
}
