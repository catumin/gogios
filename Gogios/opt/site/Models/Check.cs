using System;

namespace site.Models
{
  public class Check
    {
        public string title { get; set; }
        public string command { get; set; }
        public string expected { get; set; }
        public bool good { get; set; }
    }
  }