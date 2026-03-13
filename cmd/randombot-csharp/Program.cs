using System.Text.Json;
using System.Text.Json.Serialization;

namespace ArenaBot;

public class Program
{
    private static List<string> _axisNames = new();
    private static readonly Random _random = new();

    public static async Task Main(string[] args)
    {
        Console.Error.WriteLine("randombot-csharp: started");

        string? line;
        while ((line = await Console.In.ReadLineAsync()) != null)
        {
            try
            {
                var msg = JsonSerializer.Deserialize<ServerMessage>(line);
                if (msg == null) continue;

                switch (msg.Type)
                {
                    case "start":
                        _axisNames = msg.Axes?.Select(a => a.Name).ToList() ?? new List<string>();
                        foreach (var axis in msg.Axes ?? Enumerable.Empty<AxisDefinition>())
                        {
                            Console.Error.WriteLine($"randombot-csharp: axis registered: {axis.Name} (default: {axis.Value})");
                        }
                        break;

                    case "state":
                        var response = new InputMessage
                        {
                            Axes = new Dictionary<string, float>()
                        };
                        foreach (var name in _axisNames)
                        {
                            response.Axes[name] = (float)(_random.NextDouble() * 2 - 1);
                        }

                        var json = JsonSerializer.Serialize(response);
                        Console.WriteLine(json);
                        break;

                    case "end":
                        Console.Error.WriteLine("randombot-csharp: game ended");
                        return;
                }
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine($"randombot-csharp: error: {ex.Message}");
            }
        }
    }
}

public class ServerMessage
{
    [JsonPropertyName("type")]
    public string Type { get; set; } = "";

    [JsonPropertyName("state")]
    public JsonElement? State { get; set; }

    [JsonPropertyName("axes")]
    public List<AxisDefinition>? Axes { get; set; }
}

public class AxisDefinition
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = "";

    [JsonPropertyName("value")]
    public float Value { get; set; }
}

public class InputMessage
{
    [JsonPropertyName("axes")]
    public Dictionary<string, float> Axes { get; set; } = new();
}
