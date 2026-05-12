import { Github, Linkedin, Mail, type LucideIcon } from 'lucide-react'

export interface SocialLink {
  name: string
  url: string
  icon: LucideIcon
}

export const socialLinks: SocialLink[] = [
  {
    name: 'GitHub',
    url: 'https://github.com/madhav-poojari',
    icon: Github,
  },
  {
    name: 'LinkedIn',
    url: 'https://linkedin.com/in/madhava-p-268043209/',
    icon: Linkedin,
  },
  {
    name: 'Email',
    url: 'mailto:madhava.poojari16@gmail.com',
    icon: Mail,
  },
]
